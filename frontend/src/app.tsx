/*
 * Copyright (C) 2023. Gardel <sunxinao@hotmail.com> and contributors
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
import './app.css';
import Login from './login';
import {Container} from '@mui/material';
import {AppState} from './types';
import User from './user';
import axios from 'axios';
import {useSnackbar} from 'notistack';

function App() {
    const {enqueueSnackbar} = useSnackbar();
    const [appData, setAppData] = React.useState(() => {
        const saved = localStorage.getItem('appData');
        return (saved ? JSON.parse(saved) : {
            login: false,
            accessToken: '',
            tokenValid: false,
            loginTime: 0,
            profileName: '',
            uuid: ''
        }) as AppState;
    });

    React.useEffect(() => {
        localStorage.setItem('appData', JSON.stringify(appData));
    }, [appData]);

    const setTokenValid = (tokenValid: boolean) => appData.tokenValid != tokenValid && setAppData((oldData: AppState) => {
        return tokenValid ? {
            ...oldData,
            tokenValid: true
        } : {
            ...oldData,
            tokenValid: false,
            accessToken: '',
            loginTime: 0
        };
    });

    setTokenValid((appData.accessToken && Date.now() - appData.loginTime < 30 * 86400 * 1000) as boolean)

    if (appData.tokenValid) {
        let postData = {
            accessToken: appData.accessToken,
        };
        axios.post('/authserver/validate', postData)
            .catch(e => {
            const response = e.response;
            if (response && response.status == 403) {
                axios.post('/authserver/refresh', postData)
                    .then(response => {
                        const data = response.data;
                        if (data && data.accessToken) {
                            setAppData({
                                ...appData,
                                accessToken: data.accessToken,
                                loginTime: Date.now(),
                                profileName: data.selectedProfile?.name,
                                uuid: data.selectedProfile?.id
                            });
                            enqueueSnackbar('刷新token成功，accessToken:' + data.accessToken, {variant: 'success'});
                        } else {
                            setTokenValid(false);
                        }
                    })
                    .catch(e => {
                        const response = e.response;
                        if (response && response.status == 403) {
                            enqueueSnackbar('登录已过期', {variant: 'warning'});
                            setTokenValid(false);
                        } else {
                            enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                        }
                    });
            } else {
                enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
            }
        });
    }

    return (
        <Container maxWidth={'lg'}>
            {appData.tokenValid ? <User appData={appData} setAppData={setAppData}/> : <Login appData={appData} setAppData={setAppData}/>}
        </Container>
    );
}

export default App;
