// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhcore

type Marshaller = func(v interface{}) ([]byte, error)
type Unmarshaller = func(data []byte, v interface{}) error
