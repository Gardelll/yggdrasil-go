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

import * as THREE from 'three';
import {texturePositions} from './texture-positions';
import {BufferAttribute} from 'three';

function createCube(texture: THREE.Texture, width: number, height: number, depth: number, textures: any, slim: boolean, name: string, transparent: boolean = false) {
    let textureWidth: number = texture.image.width;
    let textureHeight: number = texture.image.height;

    let geometry = new THREE.BoxGeometry(width, height, depth);
    let material = new THREE.MeshStandardMaterial({
        /*color: 0x00ff00,*/
        map: texture,
        transparent: transparent || false,
        alphaTest: 0.1,
        side: transparent ? THREE.DoubleSide : THREE.FrontSide
    });

    geometry.computeBoundingBox();

    const uvAttribute = geometry.getAttribute('uv') as BufferAttribute;

    let faceNames = ['right', 'left', 'top', 'bottom', 'front', 'back'];
    let faceUvs = [];
    for (let i = 0; i < faceNames.length; i++) {
        let face = textures[faceNames[i]];
        // if (faceNames[i] === 'back') {
            //     console.log(face)
            // console.log("X: " + (slim && face.sx ? face.sx : face.x))
            // console.log("W: " + (slim && face.sw ? face.sw : face.w))
        // }
        let w = textureWidth;
        let h = textureHeight;
        let tx1 = ((slim && face.sx ? face.sx : face.x) / w);
        let ty1 = (face.y / h);
        let tx2 = (((slim && face.sx ? face.sx : face.x) + (slim && face.sw ? face.sw : face.w)) / w);
        let ty2 = ((face.y + face.h) / h);

        faceUvs[i] = [
            new THREE.Vector2(tx1, ty2),
            new THREE.Vector2(tx1, ty1),
            new THREE.Vector2(tx2, ty1),
            new THREE.Vector2(tx2, ty2)
        ];
        // console.log(faceUvs[i])

        let flipX = face.flipX;
        let flipY = face.flipY;

        let temp;
        if (flipY) {
            temp = faceUvs[i].slice(0);
            faceUvs[i][0] = temp[2];
            faceUvs[i][1] = temp[3];
            faceUvs[i][2] = temp[0];
            faceUvs[i][3] = temp[1];
        }
        if (flipX) {//flip x
            temp = faceUvs[i].slice(0);
            faceUvs[i][0] = temp[3];
            faceUvs[i][1] = temp[2];
            faceUvs[i][2] = temp[1];
            faceUvs[i][3] = temp[0];
        }
    }

    let j = 0;
    for (let i = 0; i < faceUvs.length; i++) {
        uvAttribute.setXY(j++, faceUvs[i][0].x, faceUvs[i][0].y);
        uvAttribute.setXY(j++, faceUvs[i][3].x, faceUvs[i][3].y);
        uvAttribute.setXY(j++, faceUvs[i][1].x, faceUvs[i][1].y);
        uvAttribute.setXY(j++, faceUvs[i][2].x, faceUvs[i][2].y);
    }
    uvAttribute.needsUpdate = true;

    let cube = new THREE.Mesh(geometry, material);
    cube.name = name;
    // cube.position.set(x, y, z);
    cube.castShadow = true;
    cube.receiveShadow = false;

    return cube;
}


export default function createPlayerModel(skinTexture: THREE.Texture, capeTexture: THREE.Texture | null | undefined, v: number, slim: boolean = false, capeType?: string): THREE.Object3D<THREE.Event> {
    let headGroup = new THREE.Object3D();
    headGroup.name = 'headGroup';
    headGroup.position.x = 0;
    headGroup.position.y = 28;
    headGroup.position.z = 0;
    headGroup.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
    let head = createCube(skinTexture,
        8, 8, 8,
        texturePositions.head[v],
        slim,
        'head'
    );
    head.translateOnAxis(new THREE.Vector3(0, 1, 0), 4);
    headGroup.add(head);
    let hat = createCube(skinTexture,
        8.667, 8.667, 8.667,
        texturePositions.hat[v],
        slim,
        'hat',
        true
    );
    hat.translateOnAxis(new THREE.Vector3(0, 1, 0), 4);
    headGroup.add(hat);

    let bodyGroup = new THREE.Object3D();
    bodyGroup.name = 'bodyGroup';
    bodyGroup.position.x = 0;
    bodyGroup.position.y = 18;
    bodyGroup.position.z = 0;
    let body = createCube(skinTexture,
        8, 12, 4,
        texturePositions.body[v],
        slim,
        'body'
    );
    bodyGroup.add(body);
    if (v >= 1) {
        let jacket = createCube(skinTexture,
            8.667, 12.667, 4.667,
            texturePositions.jacket,
            slim,
            'jacket',
            true
        );
        bodyGroup.add(jacket);
    }

    let leftArmGroup = new THREE.Object3D();
    leftArmGroup.name = 'leftArmGroup';
    leftArmGroup.position.x = slim ? -5.5 : -6;
    leftArmGroup.position.y = 18;
    leftArmGroup.position.z = 0;
    leftArmGroup.translateOnAxis(new THREE.Vector3(0, 1, 0), 4);
    let leftArm = createCube(skinTexture,
        slim ? 3 : 4, 12, 4,
        texturePositions.leftArm[v],
        slim,
        'leftArm'
    );
    leftArm.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
    leftArmGroup.add(leftArm);
    if (v >= 1) {
        let leftSleeve = createCube(skinTexture,
            slim ? 3.667 : 4.667, 12.667, 4.667,
            texturePositions.leftSleeve,
            slim,
            'leftSleeve',
            true
        );
        leftSleeve.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
        leftArmGroup.add(leftSleeve);
    }

    let rightArmGroup = new THREE.Object3D();
    rightArmGroup.name = 'rightArmGroup';
    rightArmGroup.position.x = slim ? 5.5 : 6;
    rightArmGroup.position.y = 18;
    rightArmGroup.position.z = 0;
    rightArmGroup.translateOnAxis(new THREE.Vector3(0, 1, 0), 4);
    let rightArm = createCube(skinTexture,
        slim ? 3 : 4, 12, 4,
        texturePositions.rightArm[v],
        slim,
        'rightArm'
    );
    rightArm.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
    rightArmGroup.add(rightArm);
    if (v >= 1) {
        let rightSleeve = createCube(skinTexture,
            slim ? 3.667 : 4.667, 12.667, 4.667,
            texturePositions.rightSleeve,
            slim,
            'rightSleeve',
            true
        );
        rightSleeve.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
        rightArmGroup.add(rightSleeve);
    }

    let leftLegGroup = new THREE.Object3D();
    leftLegGroup.name = 'leftLegGroup';
    leftLegGroup.position.x = -2;
    leftLegGroup.position.y = 6;
    leftLegGroup.position.z = 0;
    leftLegGroup.translateOnAxis(new THREE.Vector3(0, 1, 0), 4);
    let leftLeg = createCube(skinTexture,
        4, 12, 4,
        texturePositions.leftLeg[v],
        slim,
        'leftLeg'
    );
    leftLeg.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
    leftLegGroup.add(leftLeg);
    if (v >= 1) {
        let leftTrousers = createCube(skinTexture,
            4.667, 12.667, 4.667,
            texturePositions.leftTrousers,
            slim,
            'leftTrousers',
            true
        );
        leftTrousers.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
        leftLegGroup.add(leftTrousers);
    }

    let rightLegGroup = new THREE.Object3D();
    rightLegGroup.name = 'rightLegGroup';
    rightLegGroup.position.x = 2;
    rightLegGroup.position.y = 6;
    rightLegGroup.position.z = 0;
    rightLegGroup.translateOnAxis(new THREE.Vector3(0, 1, 0), 4);
    let rightLeg = createCube(skinTexture,
        4, 12, 4,
        texturePositions.rightLeg[v],
        slim,
        'rightLeg'
    );
    rightLeg.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
    rightLegGroup.add(rightLeg);
    if (v >= 1) {
        let rightTrousers = createCube(skinTexture,
            4.667, 12.667, 4.667,
            texturePositions.rightTrousers,
            slim,
            'rightTrousers',
            true
        );
        rightTrousers.translateOnAxis(new THREE.Vector3(0, 1, 0), -4);
        rightLegGroup.add(rightTrousers);
    }

    let playerGroup = new THREE.Object3D();
    playerGroup.add(headGroup);
    playerGroup.add(bodyGroup);
    playerGroup.add(leftArmGroup);
    playerGroup.add(rightArmGroup);
    playerGroup.add(leftLegGroup);
    playerGroup.add(rightLegGroup);

    if (capeTexture) {
        console.log(texturePositions);
        let capeTextureCoordinates = texturePositions.capeRelative;
        if (capeType === 'optifine') {
            capeTextureCoordinates = texturePositions.capeOptifineRelative;
        }
        if (capeType === 'labymod') {
            capeTextureCoordinates = texturePositions.capeLabymodRelative;
        }
        capeTextureCoordinates = JSON.parse(JSON.stringify(capeTextureCoordinates)); // bad clone to keep the below scaling from affecting everything

        console.log(capeTextureCoordinates);

        type CubeTextureKey = 'left' | 'right' | 'front' | 'back' | 'top' | 'bottom'
        // Multiply coordinates by image dimensions
        for (let cord in capeTextureCoordinates) {
            let key = cord as CubeTextureKey;
            capeTextureCoordinates[key].x *= capeTexture.image.width;
            capeTextureCoordinates[key].w *= capeTexture.image.width;
            capeTextureCoordinates[key].y *= capeTexture.image.height;
            capeTextureCoordinates[key].h *= capeTexture.image.height;
        }

        console.log(capeTextureCoordinates);

        let capeGroup = new THREE.Object3D();
        capeGroup.name = 'capeGroup';
        capeGroup.position.x = 0;
        capeGroup.position.y = 16;
        capeGroup.position.z = -2.5;
        capeGroup.translateOnAxis(new THREE.Vector3(0, 1, 0), 8);
        capeGroup.translateOnAxis(new THREE.Vector3(0, 0, 1), 0.5);
        let cape = createCube(capeTexture,
            10, 16, 1,
            capeTextureCoordinates,
            false,
            'cape');
        cape.rotation.x = toRadians(10); // slight backward angle
        cape.translateOnAxis(new THREE.Vector3(0, 1, 0), -8);
        cape.translateOnAxis(new THREE.Vector3(0, 0, 1), -0.5);
        cape.rotation.y = toRadians(180); // flip front&back to be correct
        capeGroup.add(cape);

        playerGroup.add(capeGroup);
    }

    return playerGroup;
}


function toRadians(angle: number) {
    return angle * (Math.PI / 180);
}