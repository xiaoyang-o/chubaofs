// Copyright 2018 The Chubao Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package proto

import "sync"

type UserType uint8

const (
	UserTypeInvalid UserType = 0x0
	UserTypeRoot    UserType = 0x1
	UserTypeAdmin   UserType = 0x2
	UserTypeNormal  UserType = 0x3
)

func (u UserType) Valid() bool {
	switch u {
	case UserTypeRoot,
		UserTypeAdmin,
		UserTypeNormal:
		return true
	default:
	}
	return false
}

func (u UserType) String() string {
	switch u {
	case UserTypeRoot:
		return "root"
	case UserTypeAdmin:
		return "admin"
	case UserTypeNormal:
		return "normal"
	default:
	}
	return "invalid"
}

func UserTypeFromString(name string) UserType {
	switch name {
	case "root":
		return UserTypeRoot
	case "admin":
		return UserTypeAdmin
	case "normal":
		return UserTypeNormal
	default:
	}
	return UserTypeInvalid
}

type UserAK struct {
	UserID    string `json:"user_id"`
	AccessKey string `json:"access_key"`
	Password  string `json:"password"`
}

type AKPolicy struct {
	AccessKey  string      `json:"access_key"`
	SecretKey  string      `json:"secret_key"`
	Policy     *UserPolicy `json:"policy"`
	UserID     string      `json:"user_id"`
	UserType   UserType    `json:"user_type"`
	CreateTime string      `json:"create_time"`
}

type UserPolicy struct {
	OwnVols        []string            `json:"own_vols"`
	AuthorizedVols map[string][]string `json:"authorized_vols"` // mapping: volume -> actions
	mu             sync.RWMutex
}

func NewUserPolicy() *UserPolicy {
	return &UserPolicy{
		OwnVols:        make([]string, 0),
		AuthorizedVols: make(map[string][]string),
	}
}

func NewAkPolicy() *AKPolicy {
	return &AKPolicy{Policy: NewUserPolicy()}
}

type VolAK struct {
	Vol          string              `json:"vol"`
	AKAndActions map[string][]string // k: ak, v: actions
	sync.RWMutex
}

func (policy *UserPolicy) AddOwnVol(volume string) {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	for _, ownVol := range policy.OwnVols {
		if ownVol == volume {
			return
		}
	}
	policy.OwnVols = append(policy.OwnVols, volume)
}

func (policy *UserPolicy) RemoveOwnVol(volume string) {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	for i, ownVol := range policy.OwnVols {
		if ownVol == volume {
			if i == len(policy.OwnVols)-1 {
				policy.OwnVols = policy.OwnVols[:i]
				return
			}
			policy.OwnVols = append(policy.OwnVols[:i], policy.OwnVols[i+1:]...)
			return
		}
	}
}

func (policy *UserPolicy) Add(addPolicy *UserPolicy) {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	policy.OwnVols = append(policy.OwnVols, addPolicy.OwnVols...)
	for k, v := range addPolicy.AuthorizedVols {
		if apis, ok := policy.AuthorizedVols[k]; ok {
			policy.AuthorizedVols[k] = append(apis, addPolicy.AuthorizedVols[k]...)
		} else {
			policy.AuthorizedVols[k] = v
		}
	}
}

func (policy *UserPolicy) Delete(deletePolicy *UserPolicy) {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	policy.OwnVols = removeSlice(policy.OwnVols, deletePolicy.OwnVols)
	for k, v := range deletePolicy.AuthorizedVols {
		if apis, ok := policy.AuthorizedVols[k]; ok {
			policy.AuthorizedVols[k] = removeSlice(apis, v)
		}
	}
}

func removeSlice(s []string, removeSlice []string) []string {
	if len(s) == 0 {
		return s
	}
	for _, elem := range removeSlice {
		for i, v := range s {
			if v == elem {
				s = append(s[:i], s[i+1:]...)
				break
			}
		}
	}
	return s
}

func CleanPolicy(policy *UserPolicy) (newUserPolicy *UserPolicy) {
	m := make(map[string]bool)
	newUserPolicy = NewUserPolicy()
	policy.mu.Lock()
	defer policy.mu.Unlock()
	for _, vol := range policy.OwnVols {
		if _, exist := m[vol]; !exist {
			m[vol] = true
			newUserPolicy.OwnVols = append(newUserPolicy.OwnVols, vol)
		}
	}
	for vol, apis := range policy.AuthorizedVols {
		checkMap := make(map[string]bool)
		newAPI := make([]string, 0)
		for _, api := range apis {
			if _, exist := checkMap[api]; !exist {
				checkMap[api] = true
				newAPI = append(newAPI, api)
			}
		}
		newUserPolicy.AuthorizedVols[vol] = newAPI
	}
	return
}

type UserCreateParam struct {
	ID        string `json:"id"`
	Password  string `json:"pwd"`
	AccessKey string `json:"ak"`
	SecretKey string `json:"sk"`
	Type      UserType
}

type UserUpdateParam = UserCreateParam
