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
const MDCRipple = mdc.ripple.MDCRipple;

const snackbar = new MDCSnackbar(document.querySelector(".mdc-snackbar"));
const username = new MDCTextField(document.querySelector(".username"));
const password = new MDCTextField(document.querySelector(".password"));
const profileName = new MDCTextField(document.querySelector(".profileName"));

new MDCRipple(document.querySelector(".next"));

snackbar.close();

var login = false;

$(".login").click(function (btn) {
    login = true;
    $(".profileName").hide();
    $("#profileName-input").removeAttr("required");
    $(".next").children(".mdc-button__label").text("登录");
    $(this).hide();
});

$("#reg-form").submit(function (e) {
    if (!login) {
        $.ajax({
            url: "/authserver/register",
            type: "POST",
            dataType: "JSON",
            contentType: "application/json",
            data: JSON.stringify({
                username: username.value,
                password: password.value,
                profileName: profileName.value
            }),
            success: function (data) {
                if (!data.id) {
                    if (data.errorMessage) snackbar.labelText = data.errorMessage;
                    snackbar.open();
                } else {
                    login = true;
                    $(".profileName").hide();
                    $(".login").hide();
                    $(".next").children(".mdc-button__label").text("登录");
                    snackbar.timeoutMs = 10000;
                    snackbar.labelText = "注册成功，uid:" + data.id;
                    snackbar.open();
                    localStorage.uuid = data.id;
                }
            },
            error: function (e) {
                let response = JSON.parse(e.responseText);
                if (response.errorMessage === "profileName exist") {
                    snackbar.labelText = "注册失败: 角色名已存在";
                } else if (response.errorMessage === "profileName duplicate") {
                    snackbar.labelText = "注册失败: 角色名与正版用户冲突";
                } else {
                    snackbar.labelText = "注册失败: " + response.errorMessage;
                }
                snackbar.open();
            }
        });
    } else {
        $.ajax({
            url: "/authserver/authenticate",
            type: "POST",
            dataType: "JSON",
            contentType: "application/json",
            data: JSON.stringify({
                username: username.value,
                password: password.value
            }),
            success: function (data) {
                if (!data.accessToken) {
                    snackbar.labelText = "登录失败:";
                    if (data.errorMessage) snackbar.labelText += data.errorMessage;
                    snackbar.open();
                } else {
                    snackbar.timeoutMs = 5000;
                    snackbar.labelText = "登录成功，accessToken:" + data.accessToken;
                    snackbar.open();
                    localStorage.accessToken = data.accessToken;
                    localStorage.loginTime = new Date().getTime();
                    localStorage.profileName = data.selectedProfile.name;
                    if (data.selectedProfile) {
                        localStorage.profileName = data.selectedProfile.name;
                        localStorage.uuid = data.selectedProfile.id;
                    }
                    // localStorage.username = username.value;
                    // localStorage.password = password.value;
                    setTimeout(function () {
                        window.location = "user.html";
                    }, 3000);
                }
            },
            error: function (e) {
                let response = JSON.parse(e.responseText);
                snackbar.labelText = "登录失败: " + response.errorMessage;
                snackbar.open();
            }
        });
    }

    e.preventDefault();
});

$(document).ready(function () {
    if (!localStorage.accessToken && localStorage.loginTime !== undefined &&
        (new Date().getTime() - localStorage.loginTime) < 30 * 86400 * 1000) {
        window.location = "user.html";
    }
});
