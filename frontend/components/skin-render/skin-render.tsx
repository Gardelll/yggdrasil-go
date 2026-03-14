/*
 * Copyright (C) 2023-2025. Gardel <gardel741@outlook.com> and contributors
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

import * as THREE from 'three';
import {Canvas, RootState, useFrame, useLoader} from '@react-three/fiber';
import React from 'react';
import createPlayerModel from './utils';
import {OrbitControls} from '@react-three/drei';
import {EffectComposer, SSAO} from '@react-three/postprocessing';
import {BlendFunction} from 'postprocessing';

function PlayerModel(props: { skinUrl: string, capeUrl?: string, slim?: boolean }) {
    const {skinUrl, capeUrl, slim} = props;
    const skinTexture: THREE.Texture = useLoader(THREE.TextureLoader, skinUrl);
    skinTexture.magFilter = THREE.NearestFilter;
    skinTexture.minFilter = THREE.NearestFilter;
    skinTexture.anisotropy = 0;
    skinTexture.needsUpdate = true;
    const version = skinTexture.image.height > 32 ? 1 : 0;

    const rawCapeTexture: THREE.Texture = useLoader(THREE.TextureLoader, capeUrl ?? skinUrl);
    const capeTexture = React.useMemo(() => {
        if (!capeUrl) return undefined;
        rawCapeTexture.magFilter = THREE.NearestFilter;
        rawCapeTexture.minFilter = THREE.NearestFilter;
        rawCapeTexture.anisotropy = 0;
        rawCapeTexture.needsUpdate = true;
        return rawCapeTexture;
    }, [capeUrl, rawCapeTexture]);

    const playerModel = React.useMemo(
        () => createPlayerModel(skinTexture, capeTexture, version, slim),
        [skinTexture, capeTexture, version, slim]
    );

    useFrame((_state, delta) => {
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
        <div className="h-full min-h-[300px]">

            <Canvas
                camera={{position: [0, 5, 30], near: 5}}
                gl={{antialias: true, alpha: true, preserveDrawingBuffer: true}}
                onCreated={onCanvasCreate}>
                <ambientLight color={0xffffff}/>
                <PlayerModel {...props}/>
                <OrbitControls makeDefault target={[0, 5, 0]}/>
                <EffectComposer enableNormalPass>
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
        </div>
    );
}

export default SkinRender;
