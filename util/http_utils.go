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

package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetObject(url string, value interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return &YggdrasilError{
			Status:       http.StatusNoContent,
			ErrorCode:    "IllegalArgumentException",
			ErrorMessage: "Http No Content",
		}
	} else if resp.StatusCode/100 == 4 {
		decoder := json.NewDecoder(resp.Body)
		errResp := YggdrasilError{}
		err = decoder.Decode(&errResp)
		if err != nil {
			return err
		}
		return errResp
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(body, value)
	}
}

func GetForString(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return "", nil
	} else if resp.StatusCode/100 == 4 {
		decoder := json.NewDecoder(resp.Body)
		errResp := YggdrasilError{}
		err = decoder.Decode(&errResp)
		if err != nil {
			return "", err
		}
		return "", errResp
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
}

func PostObject(url string, data interface{}, result interface{}) error {
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return &YggdrasilError{
			Status:       http.StatusNoContent,
			ErrorCode:    "IllegalArgumentException",
			ErrorMessage: "Http No Content",
		}
	} else if resp.StatusCode/100 == 4 {
		decoder := json.NewDecoder(resp.Body)
		errResp := YggdrasilError{}
		err = decoder.Decode(&errResp)
		if err != nil {
			return err
		}
		return errResp
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return json.Unmarshal(body, result)
	}
}

func PostObjectForError(url string, data interface{}) error {
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil
	} else {
		decoder := json.NewDecoder(resp.Body)
		errResp := YggdrasilError{}
		err = decoder.Decode(&errResp)
		if err != nil {
			return err
		}
		return errResp
	}
}

func PostForString(url string, data []byte) (string, error) {
	reader := bytes.NewReader(data)
	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return "", nil
	} else if resp.StatusCode/100 == 4 {
		decoder := json.NewDecoder(resp.Body)
		errResp := YggdrasilError{}
		err = decoder.Decode(&errResp)
		if err != nil {
			return "", err
		}
		return "", errResp
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
}
