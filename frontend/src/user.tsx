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
import {
    Box,
    Collapse,
    Container,
    Fade,
    FilledInput,
    FormControl,
    FormControlLabel,
    FormLabel,
    IconButton,
    InputAdornment,
    InputLabel,
    Paper,
    Radio,
    RadioGroup,
    TextField
} from '@mui/material';
import './user.css';
import {AppState} from './types';
import {Delete} from '@mui/icons-material';
import Button from '@mui/material/Button';
import {useSnackbar} from 'notistack';
import {FocusedShowHelperText} from './components';
import {SubmitHandler, useForm} from 'react-hook-form';
import axios from 'axios';
import SkinRender from './skinrender/skin-render';

function handleMouseDown(event: React.MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
}

function UploadTextureForm(props: {
    appData: AppState,
    setAppData: React.Dispatch<React.SetStateAction<AppState>>,
    skinData: SkinData | null,
    setSkinData: React.Dispatch<React.SetStateAction<SkinData | null>>
}) {
    const {appData, skinData, setSkinData} = props;
    const [submitting, setSubmitting] = React.useState(false);

    const {enqueueSnackbar} = useSnackbar();

    const fileInputElem = React.useRef<HTMLInputElement>(null);
    const [filePath, setFilePath] = React.useState('');
    const handleFilePathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setFilePath(event.target.value);

        if (skinData) {
            if (type == 'cape' && skinData.capeUrl?.startsWith('blob:')) {
                URL.revokeObjectURL(skinData.capeUrl);
            } else if (skinData.skinUrl.startsWith('blob:')) {
                URL.revokeObjectURL(skinData.skinUrl);
            }
        }

        const fileInput = event.target;
        const fileBlob = fileInput.files?.length ? fileInput.files[0] : null;
        if (fileBlob && skinData) {
            let data: SkinData = {
                ...skinData
            }
            const fakeUrl = URL.createObjectURL(fileBlob);
            if (type == 'cape') {
                data.capeUrl = fakeUrl;
            } else {
                data.slim = model == 'slim';
                data.skinUrl = fakeUrl;
            }
            setSkinData(data);
        }
    };

    const [url, setUrl] = React.useState('');
    const handleUrlChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setUrl(event.target.value);
    };

    const [type, setType] = React.useState('skin');
    const handleTypeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setType(event.target.value);
    };

    const [model, setModel] = React.useState('default');
    const handleModelChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setModel(event.target.value);

        if (skinData) {
            setSkinData({
                ...skinData,
                slim: event.target.value == 'slim'
            })
        }
    };

    const uploadTexture = (event: React.FormEvent) => {
        event.preventDefault();
        const fileInput = fileInputElem.current;
        if (fileInput == null) {
            console.error('#file-input-real is not a valid element');
            return;
        }
        setSubmitting(true);
        const fileBlob = fileInput.files?.length ? fileInput.files[0] : null;
        if (filePath && fileBlob) {
            const formData = new FormData();
            formData.append('model', model);
            formData.append('file', fileBlob);
            axios.put(`/api/user/profile/${appData.uuid}/${type}`, formData, {
                headers: {
                    'Authorization': 'Bearer ' + appData.accessToken
                }
            }).then(() => {
                enqueueSnackbar('上传成功', {variant: 'success'});
            }).catch(e => {
                const response = e.response;
                if (response && response.data) {
                    enqueueSnackbar(response.data.errorMessage, {variant: 'error'});
                } else {
                    enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                }
            }).finally(() => setSubmitting(false));
        } else if (url) {
            axios.post(`/api/user/profile/${appData.uuid}/${type}`, {model, url}, {
                headers: {
                    'Authorization': 'Bearer ' + appData.accessToken
                }
            }).then(() => {
                enqueueSnackbar('上传成功', {variant: 'success'});
                axios.get('/sessionserver/session/minecraft/profile/' + appData.uuid).then(response => {
                    let texturesProp = response.data.properties.find((v: any) => v.name == 'textures');
                    let profile: any = {};
                    if (texturesProp) {
                        profile = JSON.parse(window.atob(texturesProp.value));
                    }
                    if (profile.textures) {
                        let data: SkinData = {
                            skinUrl: ''
                        };
                        if (profile.textures.SKIN) {
                            data.skinUrl = profile.textures.SKIN.url;
                            data.slim = profile.textures.SKIN.metadata ? profile.textures.SKIN.metadata.model == 'slim' : false;
                        } else {
                            let index = getUUIDHashCode(appData.uuid) % DEFAULT_SKINS.length;
                            data.skinUrl = DEFAULT_SKINS[index].skinUrl;
                            data.slim = DEFAULT_SKINS[index].slim;
                        }
                        if (profile.textures.CAPE) {
                           data.capeUrl = profile.textures.CAPE.url;
                        }
                        setSkinData(data);
                    }
                });
            }).catch(e => {
                const response = e.response;
                if (response && response.data) {
                    enqueueSnackbar(response.data.errorMessage, {variant: 'error'});
                } else {
                    enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
                }
            }).finally(() => setSubmitting(false));
        } else {
            enqueueSnackbar('未选择文件', {variant: 'warning'});
            setSubmitting(false);
        }
    };

    const deleteTexture = () => {
        setSubmitting(true);
        axios.delete(`/api/user/profile/${appData.uuid}/${type}`, {
            headers: {
                'Authorization': 'Bearer ' + appData.accessToken
            }
        }).then(() => {
            enqueueSnackbar('删除成功', {variant: 'success'});
            if (skinData != null) {
                if (type == 'cape') {
                    setSkinData({
                        ...skinData,
                        capeUrl: undefined
                    });
                } else {
                    // 显示默认材质
                    let index = getUUIDHashCode(appData.uuid) % DEFAULT_SKINS.length;
                    setSkinData({
                        ...DEFAULT_SKINS[index],
                        capeUrl: skinData.capeUrl
                    });
                }
            }
        }).catch(e => {
            const response = e.response;
            if (response && response.data) {
                enqueueSnackbar(response.data.errorMessage, {variant: 'error'});
            } else {
                enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
            }
        }).finally(() => setSubmitting(false));
    };

    // noinspection JSUnusedGlobalSymbols
    return (
        <>
            <section className="header">
                <h3>上传材质</h3>
            </section>

            <Box component="form" autoComplete="off" onSubmit={uploadTexture}>
                <Box component="div" width="50%" boxSizing="border-box" display="inline-block">
                    <FormControl>
                        <FormLabel id="texture-type-group-label">材质类别: </FormLabel>
                        <RadioGroup
                            row
                            aria-labelledby="texture-type-group-label"
                            value={type}
                            onChange={handleTypeChange}
                            name="type">
                            <FormControlLabel value="skin" control={<Radio/>} label="皮肤"/>
                            <FormControlLabel value="cape" control={<Radio/>} label="披风"/>
                        </RadioGroup>
                    </FormControl>
                </Box>
                <Fade in={type == 'skin'}>
                    <Box component="div" width="50%" boxSizing="border-box" display="inline-block">
                        <FormControl>
                            <FormLabel id="texture-model-group-label">材质模型: </FormLabel>
                            <RadioGroup
                                row
                                aria-labelledby="texture-model-group-label"
                                value={model}
                                onChange={handleModelChange}
                                name="model">
                                <FormControlLabel value="default" control={<Radio/>} label="Steve"/>
                                <FormControlLabel value="slim" control={<Radio/>} label="Alex"/>
                            </RadioGroup>
                        </FormControl>
                    </Box>
                </Fade>
                <Collapse in={!filePath} className="url">
                    <TextField
                        id="url-input"
                        name="url"
                        fullWidth
                        label="材质 URL"
                        variant="filled"
                        required={!filePath}
                        type="url"
                        value={url}
                        onChange={handleUrlChange}
                    />
                </Collapse>
                <Collapse in={!url} className="file">
                    <FormControl fullWidth variant="filled" required={!url}>
                        <InputLabel htmlFor="file-input">或者选择一个图片</InputLabel>
                        <FilledInput
                            id="file-input"
                            required={!url}
                            endAdornment={
                                <InputAdornment position="end">
                                    <IconButton
                                        aria-label="清空选择"
                                        edge="end"
                                        onMouseDown={handleMouseDown}
                                        onClick={() => setFilePath('')}>
                                        <Delete/>
                                    </IconButton>
                                </InputAdornment>
                            }
                            value={filePath}
                            inputProps={{
                                onClick: () => fileInputElem.current?.click()
                            }}
                        />
                        <input id="file-input-real" type="file" name="file" hidden ref={fileInputElem} accept="image/*" value={filePath} onChange={handleFilePathChange}/>
                    </FormControl>
                </Collapse>
                <div className="button-container">
                    <Button variant="contained" type="submit" disabled={submitting}>上传</Button>
                    <Button variant="contained" onClick={deleteTexture} disabled={submitting}>删除材质</Button>
                </div>
            </Box>
        </>
    );
}

type ChangeProfileInputs = {
    changeTo: string
};

function ChangeProfileForm(props: { appData: AppState, setAppData: React.Dispatch<React.SetStateAction<AppState>> }) {
    const {appData, setAppData} = props;
    const setProfileName = (profileName: string) => {
        if (appData.profileName != profileName) {
            setAppData(oldData => {
                return {
                    ...oldData,
                    profileName
                };
            });
        }
    };

    const [submitting, setSubmitting] = React.useState(false);

    const {enqueueSnackbar} = useSnackbar();
    const {register, handleSubmit, formState: {errors}} = useForm<ChangeProfileInputs>();
    const onSubmit: SubmitHandler<ChangeProfileInputs> = data => {
        setSubmitting(true);
        axios.post('/authserver/change', {
            accessToken: appData.accessToken,
            changeTo: data.changeTo
        }).then(() => {
            enqueueSnackbar('更改成功', {variant: 'success'});
            setProfileName(data.changeTo);
        }).catch(e => {
            const response = e.response;
            if (response && response.data) {
                let errorMessage = response.data.errorMessage;
                let message = '更改失败: ' + errorMessage;
                if (errorMessage === 'profileName exist') {
                    message = '更改失败: 角色名已存在';
                } else if (errorMessage === 'profileName duplicate') {
                    message = '更改失败: 角色名与正版用户冲突';
                }
                enqueueSnackbar(message, {variant: 'error'});
            } else {
                enqueueSnackbar('网络错误:' + e.message, {variant: 'error'});
            }
        }).finally(() => setSubmitting(false));
    };

    return (
        <>
            <section className="header">
                <h3>更改角色名</h3>
            </section>

            <Box component="form" autoComplete="off" onSubmit={handleSubmit(onSubmit)}>
                <div className="changeTo">
                    <FormControl fullWidth variant="filled" required error={errors.changeTo != null}>
                        <InputLabel htmlFor="changeTo-input">角色名</InputLabel>
                        <FilledInput
                            id="changeTo-input"
                            name="changeTo"
                            defaultValue={appData.profileName}
                            required
                            inputProps={{
                                minLength: '2', maxLength: 16,
                                ...register('changeTo', {required: true, minLength: 2, pattern: /^[a-zA-Z0-9_]{1,16}$/, maxLength: 16})
                            }}
                        />
                        <FocusedShowHelperText id="profileName-input-helper-text">字母，数字或下划线</FocusedShowHelperText>
                    </FormControl>
                </div>
                <div className="button-container">
                    <Button variant="contained" type="submit" disabled={submitting}>更改</Button>
                </div>
            </Box>
        </>
    );
}

type SkinData = {
    skinUrl: string
    capeUrl?: string
    slim?: boolean
}

const DEFAULT_SKINS: SkinData[] = [
    {skinUrl: 'player/slim/alex.png', slim: true},
    {skinUrl: 'player/slim/ari.png', slim: true},
    {skinUrl: 'player/slim/efe.png', slim: true},
    {skinUrl: 'player/slim/kai.png', slim: true},
    {skinUrl: 'player/slim/makena.png', slim: true},
    {skinUrl: 'player/slim/noor.png', slim: true},
    {skinUrl: 'player/slim/steve.png', slim: true},
    {skinUrl: 'player/slim/sunny.png', slim: true},
    {skinUrl: 'player/slim/zuri.png', slim: true},

    {skinUrl: 'player/wide/alex.png'},
    {skinUrl: 'player/wide/ari.png'},
    {skinUrl: 'player/wide/efe.png'},
    {skinUrl: 'player/wide/kai.png'},
    {skinUrl: 'player/wide/makena.png'},
    {skinUrl: 'player/wide/noor.png'},
    {skinUrl: 'player/wide/steve.png'},
    {skinUrl: 'player/wide/sunny.png'},
    {skinUrl: 'player/wide/zuri.png'},
];

function User(props: { appData: AppState, setAppData: React.Dispatch<React.SetStateAction<AppState>> }) {
    const {appData, setAppData} = props;

    const [skinData, setSkinData] = React.useState<SkinData | null>(null);

    React.useEffect(() => {
        setSkinData(null);
        axios.get('/sessionserver/session/minecraft/profile/' + appData.uuid).then(response => {
            let texturesProp = response.data.properties.find((v: any) => v.name == 'textures');
            let profile: any = {};
            if (texturesProp) {
                profile = JSON.parse(window.atob(texturesProp.value));
            }
            if (profile.textures && profile.textures.SKIN) {
                let skinUrl = profile.textures.SKIN.url;
                let slim = profile.textures.SKIN.metadata ? profile.textures.SKIN.metadata.model == 'slim' : false;
                let capeUrl = undefined;
                if (profile.textures.CAPE) {
                    capeUrl = profile.textures.CAPE.url;
                }
                setSkinData({
                    skinUrl,
                    capeUrl,
                    slim
                });
            } else if (profile.textures && profile.textures.CAPE) {
                // 显示默认材质
                let index = getUUIDHashCode(appData.uuid) % DEFAULT_SKINS.length;
                setSkinData({
                    ...DEFAULT_SKINS[index],
                    capeUrl: profile.textures.CAPE.url
                });
            } else {
                // 显示默认材质
                let index = getUUIDHashCode(appData.uuid) % DEFAULT_SKINS.length;
                setSkinData(DEFAULT_SKINS[index]);
            }
        });
    }, [appData]);

    return (
        <Container maxWidth={'sm'}>
            <Paper className={'user-card'}>
                <section className="header">
                    <h1>简陋信息页</h1>
                </section>

                <UploadTextureForm appData={appData} setAppData={setAppData} skinData={skinData} setSkinData={setSkinData}/>

                {skinData && <SkinRender skinUrl={skinData.skinUrl} capeUrl={skinData.capeUrl} slim={skinData.slim}/>}

                <ChangeProfileForm appData={appData} setAppData={setAppData}/>
            </Paper>
        </Container>
    );
}

function getUUIDHashCode(uuid: string) {
    const uuidNoDash = uuid.replace(/-/g, '');
    const mostMost = parseInt(uuidNoDash.substring(0, 8), 16);
    const mostLeast = parseInt(uuidNoDash.substring(8, 16), 16);
    const leastMost = parseInt(uuidNoDash.substring(16, 24), 16);
    const leastLeast = parseInt(uuidNoDash.substring(24, 32), 16);
    return mostMost ^ mostLeast ^ leastMost ^ leastLeast;
}

export default User;