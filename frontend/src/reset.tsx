/*
 * Copyright (C) 2023-2025. Gardel <sunxinao@hotmail.com> and contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import React from 'react';
import Button from '@mui/material/Button';
import {
    Box,
    Container,
    FilledInput,
    FormControl,
    IconButton,
    InputAdornment,
    InputLabel,
    Paper,
    TextField
} from '@mui/material';
import {Visibility, VisibilityOff} from '@mui/icons-material';
import {AppState} from './types';
import './reset.css';
import {SubmitHandler, useForm} from 'react-hook-form';
import axios from 'axios';
import {useSnackbar} from 'notistack';

type Inputs = {
    username: string,
    profileName: string,
    password: string
};

function PasswordReset(props: { appData: AppState, setAppData: React.Dispatch<React.SetStateAction<AppState>> }) {
    const {appData, setAppData} = props;
    const {enqueueSnackbar} = useSnackbar();
    const {register, handleSubmit, formState: {errors}} = useForm<Inputs>();
    const [submitting, setSubmitting] = React.useState(false);
    const onSubmit: SubmitHandler<Inputs> = data => {
        setSubmitting(true);
        const hash = window.location.hash;
        if (!hash) {
            setSubmitting(false);
            enqueueSnackbar('链接失效，请重新打开', {variant: 'error'});
            return;
        }
        const params = new URLSearchParams(hash.substring(1));
        axios.post('/authserver/resetPassword', {
            email: data.username,
            password: data.password,
            accessToken: params.get('passwordResetToken'),
        })
            .then(() => {
                toLogin();
                window.location.replace('/profile/')
                enqueueSnackbar("重置成功", {variant: 'success'});
            })
            .catch(e => {
                const response = e.response;
                if (response && response.status >= 400 && response.status < 500) {
                    let errorMessage = response.data.errorMessage ?? response.data;
                    enqueueSnackbar('重置失败: ' + errorMessage, {variant: 'error'});
                } else {
                    enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                }
            })
            .finally(() => setSubmitting(false))
    };

    const [showPassword, setShowPassword] = React.useState(false);

    const handleClickShowPassword = () => setShowPassword((show) => !show);

    const handleMouseDownPassword = (event: React.MouseEvent<HTMLButtonElement>) => {
        event.preventDefault();
    };

    const toLogin = () => {
        setAppData({
            ...appData,
            passwordReset: false
        });
    }

    return (
        <Container maxWidth={'sm'}>
            <Paper className={'reset-card'}>
                <section className="header">
                    <h1>简陋重置密码页</h1>
                </section>
                <Box component="form" autoComplete="off" onSubmit={handleSubmit(onSubmit)}>
                    <div className='username'>
                        <TextField
                            id="username-input"
                            name='username'
                            fullWidth
                            label="邮箱"
                            variant="filled"
                            required
                            error={errors.username && true}
                            type='email'
                            slotProps={{
                                htmlInput: {
                                    ...register('username', {required: true})
                                }
                            }}
                        />
                    </div>
                    <div className='password'>
                        <FormControl fullWidth variant="filled" required error={errors.password && true}>
                            <InputLabel htmlFor="password-input">新密码</InputLabel>
                            <FilledInput
                                id="password-input"
                                name="password"
                                required
                                type={showPassword ? 'text' : 'password'}
                                endAdornment={
                                    <InputAdornment position="end">
                                        <IconButton
                                            aria-label="显示密码"
                                            onClick={handleClickShowPassword}
                                            onMouseDown={handleMouseDownPassword}
                                            edge="end">
                                            {showPassword ? <VisibilityOff/> : <Visibility/>}
                                        </IconButton>
                                    </InputAdornment>
                                }
                                inputProps={{
                                    minLength: '6',
                                    ...register('password', {required: true, minLength: 6})
                                }}
                            />
                        </FormControl>
                    </div>
                    <div className='button-container'>
                        <Button variant='contained' onClick={toLogin}>登录</Button>
                        <Button variant='contained' type='submit' disabled={submitting}>重置</Button>
                    </div>
                </Box>
            </Paper>
        </Container>
    );
}

export default PasswordReset;