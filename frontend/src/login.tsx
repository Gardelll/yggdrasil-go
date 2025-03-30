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
    Collapse,
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
import './login.css';
import {SubmitHandler, useForm} from 'react-hook-form';
import axios from 'axios';
import {useSnackbar} from 'notistack';
import {FocusedShowHelperText} from './components';

type Inputs = {
    username: string,
    profileName: string,
    password: string
};

function Login(props: { appData: AppState, setAppData: React.Dispatch<React.SetStateAction<AppState>> }) {
    const {appData, setAppData} = props;
    const {enqueueSnackbar} = useSnackbar();
    const {register, handleSubmit, formState: {errors}} = useForm<Inputs>();
    const [submitting, setSubmitting] = React.useState(false);
    const [submitText, setSubmitText] = React.useState('提交');
    React.useEffect(() => {
        if (submitting) {
            return
        }
        switch (appData.login) {
            case 'login':
                setSubmitText('登录');
                break;
            case 'register':
                setSubmitText('注册');
                break;
            case 'reset':
                setSubmitText('重置');
                break;
            default:
                setSubmitText('提交');
                break;
        }
    }, [appData, submitting])
    const onSubmit: SubmitHandler<Inputs> = data => {
        setSubmitting(true);
        if (appData.login === 'login') {
            axios.post('/authserver/authenticate', {
                username: data.username,
                password: data.password
            })
                .then(response => {
                    let data = response.data
                    if (data && data.accessToken) {
                        enqueueSnackbar("登录成功，accessToken:" + data.accessToken, {variant: 'success'});
                        setAppData({
                            ...appData,
                            accessToken: data.accessToken,
                            tokenValid: true,
                            loginTime: Date.now(),
                            profileName: data.selectedProfile?.name,
                            uuid: data.selectedProfile?.id
                        });
                    } else {
                        enqueueSnackbar(data && data.errorMessage ? "登录失败: " + data.errorMessage: "登陆失败", {variant: 'error'});
                    }
                })
                .catch(e => {
                    const response = e.response;
                    if (response && response.status == 403) {
                        enqueueSnackbar('登录失败: ' + response.data.errorMessage, {variant: 'error'});
                    } else {
                        enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                    }
                })
                .finally(() => setSubmitting(false))
        } else if (appData.login === 'register') {
            axios.post('/authserver/register', {
                username: data.username,
                password: data.password,
                profileName: data.profileName
            })
                .then(response => {
                    let data = response.data
                    if (data && data.id) {
                        enqueueSnackbar("注册成功，uuid:" + data.id, {variant: 'success'});
                        setLogin('login')
                    } else {
                        enqueueSnackbar(data && data.errorMessage ? "注册失败: " + data.errorMessage: "注册失败", {variant: 'error'});
                    }
                })
                .catch(e => {
                    const response = e.response;
                    if (response && response.data) {
                        let errorMessage = response.data.errorMessage ?? response.data;
                        let message =  "注册失败: " + errorMessage;
                        if (errorMessage === "profileName exist") {
                            message = "注册失败: 角色名已存在";
                        } else if (errorMessage === "profileName duplicate") {
                            message = "注册失败: 角色名与正版用户冲突";
                        }
                        enqueueSnackbar(message, {variant: 'error'});
                    } else {
                        enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                    }
                })
                .finally(() => setSubmitting(false))
        } else if (appData.login === 'reset') {
            const countdown = {
                timeout: 60,
            }
            setSubmitText(`${countdown.timeout}`);
            const timer = setInterval(() => {
                countdown.timeout = countdown.timeout - 1;
                if (countdown.timeout <= 0) {
                    clearInterval(timer);
                    setSubmitting(false);
                    return
                }
                setSubmitting(true);
                setSubmitText(`${countdown.timeout}`);
            }, 1000);
            axios.post('/authserver/sendEmail', {
                email: data.username,
                emailType: "resetPassword"
            })
                .then(() => {
                    enqueueSnackbar("重置链接发送成功，请检查垃圾邮箱", {variant: 'success'});
                })
                .catch(e => {
                    const response = e.response;
                    if (response && response.data) {
                        let errorMessage = response.data.errorMessage ?? response.data;
                        enqueueSnackbar("发送失败: " + errorMessage, {variant: 'error'});
                    } else {
                        enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                    }
                    countdown.timeout = 0;
                });
        }
    };

    const [showPassword, setShowPassword] = React.useState(false);

    const handleClickShowPassword = () => setShowPassword((show) => !show);

    const handleMouseDownPassword = (event: React.MouseEvent<HTMLButtonElement>) => {
        event.preventDefault();
    };

    const setLogin = (login: string) => setAppData((oldData: AppState) => {
        return {
            ...oldData,
            login
        };
    });

    return (
        <Container maxWidth={'sm'}>
            <Paper className={'login-card'}>
                <section className="header">
                    <h1>简陋注册页</h1>
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
                    <Collapse in={appData.login === 'register'} className='profileName'>
                        <FormControl fullWidth variant="filled" required={appData.login === 'register'} error={errors.profileName && true}>
                            <InputLabel htmlFor="profileName-input">角色名</InputLabel>
                            <FilledInput
                                id="profileName-input"
                                name="profileName"
                                required={appData.login === 'register'}
                                inputProps={appData.login !== 'register' ? {} : {
                                    minLength: '2', maxLength: 16,
                                    ...register('profileName', {required: true, minLength: 2, pattern: /^[a-zA-Z0-9_]{1,16}$/, maxLength: 16})
                                }}
                            />
                            <FocusedShowHelperText id="profileName-input-helper-text">字母，数字或下划线</FocusedShowHelperText>
                        </FormControl>
                    </Collapse>
                    <Collapse in={appData.login !== 'reset'} className='password'>
                        <FormControl fullWidth variant="filled" required={appData.login !== 'reset'} error={errors.password && true}>
                            <InputLabel htmlFor="password-input">密码</InputLabel>
                            <FilledInput
                                id="password-input"
                                name="password"
                                required={appData.login !== 'reset'}
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
                                inputProps={appData.login === 'reset' ? {} : {
                                    minLength: '6',
                                    ...register('password', {required: true, minLength: 6})
                                }}
                            />
                            <FocusedShowHelperText id="password-input-helper-text">请妥善保管密码</FocusedShowHelperText>
                        </FormControl>
                    </Collapse>
                    <div className='button-container'>
                        {appData.login !== 'reset' && <Button variant='contained' onClick={() => setLogin('reset')}>忘记密码</Button>}
                        {appData.login !== 'login' && <Button variant='contained' onClick={() => setLogin('login')}>已有账号登录</Button>}
                        {appData.login !== 'register' && <Button variant='contained' onClick={() => setLogin('register')}>注册</Button>}
                        <Button variant='contained' type='submit' disabled={submitting}>{submitText}</Button>
                    </div>
                </Box>
            </Paper>
        </Container>
    );
}

export default Login;