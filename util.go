// Copyright (c) 2026 上海飞函安全科技有限公司 (Shanghai Feihan Security Technology Co., Ltd.)
// SPDX-License-Identifier: Apache-2.0

package fhsdk

import "encoding/json"

func Pretty(obj interface{}) string {
	s, _ := json.MarshalIndent(obj, "", "  ")
	return string(s)
}
