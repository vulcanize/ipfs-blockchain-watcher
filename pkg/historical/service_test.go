// VulcanizeDB
// Copyright © 2019 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package historical_test

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/statediff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipld-eth-indexer/pkg/eth"
	"github.com/vulcanize/ipld-eth-indexer/pkg/eth/mocks"
	"github.com/vulcanize/ipld-eth-indexer/pkg/historical"
	"github.com/vulcanize/ipld-eth-indexer/pkg/shared"
)

var _ = Describe("BackFiller", func() {
	Describe("FillGaps", func() {
		It("Periodically checks for and fills in gaps in the watcher's data", func() {
			mockTransformer := &mocks.IterativeTransformer{
				ReturnErr:     nil,
				ReturnHeights: []uint64{100, 101},
			}
			mockRetriever := &mocks.Retriever{
				FirstBlockNumberToReturn: 0,
				GapsToRetrieve: []eth.DBGap{
					{
						Start: 100, Stop: 101,
					},
				},
			}
			mockFetcher := &mocks.PayloadFetcher{
				PayloadsToReturn: map[uint64]statediff.Payload{
					100: mocks.MockStateDiffPayload,
					101: mocks.MockStateDiffPayload,
				},
			}
			quitChan := make(chan bool, 1)
			backfiller := &historical.Service{
				Transformer:       mockTransformer,
				Fetcher:           mockFetcher,
				Retriever:         mockRetriever,
				GapCheckFrequency: time.Second * 2,
				BatchSize:         shared.DefaultMaxBatchSize,
				Workers:           shared.DefaultMaxBatchNumber,
				QuitChan:          quitChan,
			}
			wg := &sync.WaitGroup{}
			backfiller.Sync(wg)
			time.Sleep(time.Second * 3)
			quitChan <- true
			Expect(len(mockTransformer.PassedStateDiffs)).To(Equal(2))
			Expect(mockTransformer.PassedStateDiffs[0]).To(Equal(mocks.MockStateDiffPayload))
			Expect(mockTransformer.PassedStateDiffs[1]).To(Equal(mocks.MockStateDiffPayload))
			Expect(mockRetriever.CalledTimes).To(Equal(1))
			Expect(len(mockFetcher.CalledAtBlockHeights)).To(Equal(1))
			Expect(mockFetcher.CalledAtBlockHeights[0]).To(Equal([]uint64{100, 101}))
		})

		It("Works for single block `ranges`", func() {
			mockTransformer := &mocks.IterativeTransformer{
				ReturnErr:     nil,
				ReturnHeights: []uint64{100},
			}
			mockRetriever := &mocks.Retriever{
				FirstBlockNumberToReturn: 0,
				GapsToRetrieve: []eth.DBGap{
					{
						Start: 100, Stop: 100,
					},
				},
			}
			mockFetcher := &mocks.PayloadFetcher{
				PayloadsToReturn: map[uint64]statediff.Payload{
					100: mocks.MockStateDiffPayload,
				},
			}
			quitChan := make(chan bool, 1)
			backfiller := &historical.Service{
				Transformer:       mockTransformer,
				Fetcher:           mockFetcher,
				Retriever:         mockRetriever,
				GapCheckFrequency: time.Second * 2,
				BatchSize:         shared.DefaultMaxBatchSize,
				Workers:           shared.DefaultMaxBatchNumber,
				QuitChan:          quitChan,
			}
			wg := &sync.WaitGroup{}
			backfiller.Sync(wg)
			time.Sleep(time.Second * 3)
			quitChan <- true
			Expect(len(mockTransformer.PassedStateDiffs)).To(Equal(1))
			Expect(mockTransformer.PassedStateDiffs[0]).To(Equal(mocks.MockStateDiffPayload))
			Expect(mockRetriever.CalledTimes).To(Equal(1))
			Expect(len(mockFetcher.CalledAtBlockHeights)).To(Equal(1))
			Expect(mockFetcher.CalledAtBlockHeights[0]).To(Equal([]uint64{100}))
		})

		It("Finds beginning gap", func() {
			mockTransformer := &mocks.IterativeTransformer{
				ReturnErr:     nil,
				ReturnHeights: []uint64{0, 1, 2},
			}
			mockRetriever := &mocks.Retriever{
				FirstBlockNumberToReturn: 3,
				GapsToRetrieve: []eth.DBGap{
					{
						Start: 0,
						Stop:  2,
					},
				},
			}
			mockFetcher := &mocks.PayloadFetcher{
				PayloadsToReturn: map[uint64]statediff.Payload{
					0: mocks.MockStateDiffPayload,
					1: mocks.MockStateDiffPayload,
					2: mocks.MockStateDiffPayload,
				},
			}
			quitChan := make(chan bool, 1)
			backfiller := &historical.Service{
				Transformer:       mockTransformer,
				Fetcher:           mockFetcher,
				Retriever:         mockRetriever,
				GapCheckFrequency: time.Second * 2,
				BatchSize:         shared.DefaultMaxBatchSize,
				Workers:           shared.DefaultMaxBatchNumber,
				QuitChan:          quitChan,
			}
			wg := &sync.WaitGroup{}
			backfiller.Sync(wg)
			time.Sleep(time.Second * 3)
			quitChan <- true
			Expect(len(mockTransformer.PassedStateDiffs)).To(Equal(3))
			Expect(mockTransformer.PassedStateDiffs[0]).To(Equal(mocks.MockStateDiffPayload))
			Expect(mockTransformer.PassedStateDiffs[1]).To(Equal(mocks.MockStateDiffPayload))
			Expect(mockTransformer.PassedStateDiffs[2]).To(Equal(mocks.MockStateDiffPayload))
			Expect(mockRetriever.CalledTimes).To(Equal(1))
			Expect(len(mockFetcher.CalledAtBlockHeights)).To(Equal(1))
			Expect(mockFetcher.CalledAtBlockHeights[0]).To(Equal([]uint64{0, 1, 2}))
		})
	})
})
