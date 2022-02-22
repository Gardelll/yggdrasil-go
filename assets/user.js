/*
 * Copyright (C) 2022. Gardel <sunxinao@hotmail.com> and contributors
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

const MDCSnackbar = mdc.snackbar.MDCSnackbar;
const MDCTextField = mdc.textField.MDCTextField;
const MDCFormField = mdc.formField.MDCFormField;
const MDCRadio = mdc.radio.MDCRadio;
//const MDCRipple = mdc.ripple.MDCRipple;

const snackbar = new MDCSnackbar(document.querySelector('.mdc-snackbar'));
const modelField = new MDCFormField(document.querySelector('.model'));
const radio1 = new MDCRadio(document.querySelector('#radio-steve'));
const radio2 = new MDCRadio(document.querySelector('#radio-alex'));
const radio3 = new MDCRadio(document.querySelector('#radio-skin'));
const radio4 = new MDCRadio(document.querySelector('#radio-cape'));
modelField.input = {radio1, radio2};
const url = new MDCTextField(document.querySelector('.url'));
const file = new MDCTextField(document.querySelector('.file'));
const changeTo = new MDCTextField(document.querySelector('.changeTo'));

const modelTypeForm = document.querySelector('.mdc-form-field.model');

snackbar.close()

$("#file-input").change(function() {
    const path = this.value;
    if (!path) {
        $("#file-path").text("或选择一张图片");
        $(".url").show();
    } else {
        $("#file-path").text(path);
        url.value = '';
        $(".url").hide();
    }
});

const modelTypeChange = function () {
    if (radio3.checked) {
        $(modelTypeForm).show();
    } else {
        $(modelTypeForm).hide();
    }
}
$("#radio-3").change(modelTypeChange)
$("#radio-4").change(modelTypeChange)

$("#url-input").on("input", function() {
    const url = this.value;
    if (!url) {
        $(".file").show();
    } else {
        file.value = '';
        $(".file").hide();
    }
});

$(document).ready(function() {
    if (!localStorage.accessToken) {
        localStorage.loginTime = 1;
        window.location = "index.html";
    }
    $.ajax({
        url: '/authserver/validate',
        type: 'POST',
        dataType: "JSON",
        contentType: "application/json",
        data: JSON.stringify({
            accessToken: localStorage.accessToken,
        }),
        success: function(data) {
            //有效，啥也不整
        },
        error: function(e) {
            if (e.status == 403) {
                // 持续套娃
                $.ajax({
                    url: '/authserver/refresh',
                    type: 'POST',
                    dataType: "JSON",
                    contentType: "application/json",
                    data: JSON.stringify({
                        accessToken: localStorage.accessToken,
                    }),
                    success: function(data) {
                        if (!data.accessToken) {
                            localStorage.loginTime = 1;
                            window.location = "index.html";
                        } else {
                            snackbar.timeoutMs = 5000;
                            snackbar.labelText = "刷新token成功，accessToken:" + data.accessToken;
                            snackbar.open();
                            localStorage.accessToken = data.accessToken;
                            localStorage.loginTime = new Date().getTime();
                            if(data.selectedProfile) {
                                localStorage.profileName = data.selectedProfile.name;
                                localStorage.uuid = data.selectedProfile.id;
                            }
                        }
                    },
                    error: function(e) {
                        if (e.status == 403) {
                            localStorage.loginTime = 1;
                            window.location = "index.html";
                        }
                    }
                });
            }
        }
    });
});

$("#upload-form").submit(function(e) {
    e.preventDefault();
    if (!url.value && !$("#file-input").val()) {
        snackbar.timeoutMs = 5000;
        snackbar.labelText = "没填信息";
        snackbar.open();
        return;
    }
    let textureType = 'skin'
    if (radio3.checked) {
        textureType = 'skin'
    } else if (radio4.checked) {
        textureType = 'cape'
    }
    if (!url.value) {
        const formData = new FormData();
        formData.append("model", radio1.checked ? radio1.value : radio2.value);
        formData.append("file", $("#file-input")[0].files[0]);
        //formData.contentType = "multipart/form-data";
        $.ajax({
            url: `/api/user/profile/${localStorage.uuid}/${textureType}`,
            type: 'PUT',
            processData: false,
            contentType: false,
            headers: {'Authorization':'Bearer ' + localStorage.accessToken},
            data: formData,
            success: function(data) {
                snackbar.timeoutMs = 5000;
                snackbar.labelText = "材质上传成功";
                snackbar.open();
            },
            error: function(e) {
                snackbar.timeoutMs = 5000;
                snackbar.labelText = "材质上传失败";
                snackbar.open();
            }
        });
    } else if (url.value) {
        $.ajax({
            url: `/api/user/profile/${localStorage.uuid}/${textureType}`,
            type: 'POST',
            dataType: "JSON",
            contentType: "application/json",
            headers: {'Authorization':'Bearer ' + localStorage.accessToken},
            data: JSON.stringify({
                model: radio1.checked ? radio1.value : radio2.value,
                url: url.value
            }),
            success: function(data) {
                snackbar.timeoutMs = 5000;
                snackbar.labelText = "材质上传成功";
                snackbar.open();
            },
            error: function(e) {
                snackbar.timeoutMs = 5000;
                snackbar.labelText = "材质上传失败";
                snackbar.open();
            }
        });
    }
});

$("#change-form").submit(function(e) {
    e.preventDefault();
    if (changeTo.value.length <= 1) {
        snackbar.timeoutMs = 5000;
        snackbar.labelText = "更改失败, 角色名格式不正确";
        snackbar.open();
        return;
    }
    $.ajax({
        url: '/authserver/change',
        type: 'POST',
        dataType: "JSON",
        contentType: "application/json",
        data: JSON.stringify({
            accessToken: localStorage.accessToken,
            changeTo: changeTo.value
        }),
        success: function(data) {
            snackbar.timeoutMs = 5000;
            snackbar.labelText = "更改成功";
            snackbar.open();
            localStorage.profileName = changeTo.value;
        },
        error: function(e) {
            snackbar.timeoutMs = 5000;
            snackbar.labelText = "更改失败, 可能是角色名已存在";
            snackbar.open();
        }
    });
});

$('#delete-btn').click(function () {
    let textureType = 'skin'
    if (radio3.checked) {
        textureType = 'skin'
    } else if (radio4.checked) {
        textureType = 'cape'
    }
    $.ajax({
        url: `/api/user/profile/${localStorage.uuid}/${textureType}`,
        type: 'DELETE',
        headers: {'Authorization':'Bearer ' + localStorage.accessToken},
        success: function(data) {
            snackbar.timeoutMs = 5000;
            snackbar.labelText = "恢复成功";
            snackbar.open();
            localStorage.profileName = changeTo.value;
        },
        error: function(e) {
            snackbar.timeoutMs = 5000;
            snackbar.labelText = "重置失败";
            snackbar.open();
        }
    });
})