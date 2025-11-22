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

import {Box} from '@mui/material';
import * as THREE from 'three';
import {Canvas, RootState, useFrame, useLoader} from '@react-three/fiber';
import React from 'react';
import createPlayerModel from './utils';
import {OrbitControls} from '@react-three/drei';
import {EffectComposer, SSAO} from '@react-three/postprocessing';
import {BlendFunction} from 'postprocessing';

function PlayerModel(props: { skinUrl: string, capeUrl?: string, slim?: boolean }) {
    const {skinUrl, capeUrl, slim} = props;
    console.log(props);
    const skinTexture: THREE.Texture = useLoader(THREE.TextureLoader, skinUrl);
    skinTexture.magFilter = THREE.NearestFilter;
    skinTexture.minFilter = THREE.NearestFilter;
    skinTexture.anisotropy = 0;
    skinTexture.needsUpdate = true;
    let version = 0;
    if (skinTexture.image.height > 32) {
        version = 1;
    }
    let capeTexture: THREE.Texture | undefined = undefined;
    if (capeUrl) {
        capeTexture = useLoader(THREE.TextureLoader, capeUrl);
        if (capeTexture) {
            capeTexture.magFilter = THREE.NearestFilter;
            capeTexture.minFilter = THREE.NearestFilter;
            capeTexture.anisotropy = 0;
            capeTexture.needsUpdate = true;
        }
    }
    let playerModel = createPlayerModel(skinTexture, capeTexture, version, slim);
    useFrame((state, delta) => {
        playerModel.rotation.y += delta * 0.7;
    });
    return (
        <primitive object={playerModel} position={[0, -10, 0]}/>
    );
}

function SkinRender(props: { skinUrl: string, capeUrl?: string, slim?: boolean }) {
    const onCanvasCreate = (state: RootState) => {
        state.gl.shadowMap.enabled = true;
        state.gl.shadowMap.type = THREE.PCFSoftShadowMap;
    };
    return (
        <Box component="div" height="600px">
            <section className="header">
                <h3>预览</h3>
            </section>

            <Canvas
                camera={{position: [0, 15, 35], near: 5}}
                gl={{antialias: true, alpha: true, preserveDrawingBuffer: true}}
                onCreated={onCanvasCreate}>
                <ambientLight color={0xffffff}/>
                <PlayerModel {...props}/>
                <OrbitControls makeDefault/>
                <EffectComposer>
                    <SSAO
                        blendFunction={BlendFunction.OVERLAY}
                        samples={30}
                        rings={4}
                        distanceThreshold={1.0}
                        distanceFalloff={0.0}
                        rangeThreshold={0.5}
                        rangeFalloff={0.1}
                        luminanceInfluence={0.9}
                        radius={20}
                        resolutionScale={0.5}
                        bias={0.5}
                    />
                </EffectComposer>
            </Canvas>
        </Box>
    );
}

export default SkinRender;