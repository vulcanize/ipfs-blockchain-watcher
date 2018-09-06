// Copyright 2018 Vulcanize
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package debt_ceiling

import (
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/vulcanize/vulcanizedb/pkg/transformers/pit_file/debt_ceiling"
	"github.com/vulcanize/vulcanizedb/pkg/transformers/test_data"
)

type MockPitFileDebtCeilingConverter struct {
	converterErr          error
	PassedContractAddress string
	PassedContractABI     string
	PassedLog             types.Log
}

func (converter *MockPitFileDebtCeilingConverter) ToModel(contractAddress string, contractAbi string, ethLog types.Log) (debt_ceiling.PitFileDebtCeilingModel, error) {
	converter.PassedContractAddress = contractAddress
	converter.PassedContractABI = contractAbi
	converter.PassedLog = ethLog
	return test_data.PitFileDebtCeilingModel, converter.converterErr
}

func (converter *MockPitFileDebtCeilingConverter) SetConverterError(e error) {
	converter.converterErr = e
}