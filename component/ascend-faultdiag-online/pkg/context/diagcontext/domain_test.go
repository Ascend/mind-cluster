/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package diagcontext some test case for the domain.
*/
package diagcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var factory = NewDomainFactory()

func TestDomain_GetDomainKey(t *testing.T) {
	// 指标域的type1:指标域的name1-指标域的type2:指标域的name2
	assert.Equal(t, domain.GetDomainKey(), "domain_type_string:domain_item_1-domain_type_string:domain_item_2")
}

func TestDomain_Size(t *testing.T) {
	domainLen := 2
	assert.Equal(t, len(domain.DomainItems), domainLen)
	assert.Equal(t, domain.Size(), domainLen)
}

func TestNewDomainFactory(t *testing.T) {
	assert.NotNil(t, factory)
}

func TestDomainFactory_GetInstance(t *testing.T) {
	key := buildDomainItemsKey(domainItems)
	_, notExit := factory.domainMap[key]
	assert.False(t, notExit)

	domainCrate := factory.GetInstance(domainItems)
	assert.Equal(t, domainCrate, domain, "实例不存在时，通过传入的domainItems创建domain实例")

	_, ok := factory.domainMap[key]
	assert.True(t, ok)
	domainInstance := factory.GetInstance(domainItems)
	assert.Equal(t, domainInstance, domain, "实例存在时，直接返回该实例")
}
