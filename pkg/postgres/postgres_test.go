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

package postgres_test

import (
	"fmt"
	"strings"

	"github.com/vulcanize/ipld-eth-indexer/pkg/shared"

	"math/big"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipld-eth-indexer/pkg/node"
	"github.com/vulcanize/ipld-eth-indexer/pkg/postgres"
)

var _ = Describe("Postgres DB", func() {
	var (
		db  *postgres.DB
		err error
	)
	BeforeEach(func() {
		db, err = shared.SetupDB()
		Expect(err).ToNot(HaveOccurred())
	})

	It("serializes big.Int to db", func() {
		// postgres driver doesn't support go big.Int type
		// various casts in golang uint64, int64, overflow for
		// transaction value (in wei) even though
		// postgres numeric can handle an arbitrary
		// sized int, so use string representation of big.Int
		// and cast on insert

		bi := new(big.Int)
		bi.SetString("34940183920000000000", 10)
		Expect(bi.String()).To(Equal("34940183920000000000"))

		defer db.Exec(`DROP TABLE IF EXISTS example`)
		_, err = db.Exec("CREATE TABLE example ( id INTEGER, data NUMERIC )")
		Expect(err).ToNot(HaveOccurred())

		sqlStatement := `  
			INSERT INTO example (id, data)
			VALUES (1, cast($1 AS NUMERIC))`
		_, err = db.Exec(sqlStatement, bi.String())
		Expect(err).ToNot(HaveOccurred())

		var data string
		err = db.QueryRow(`SELECT data FROM example WHERE id = 1`).Scan(&data)
		Expect(err).ToNot(HaveOccurred())

		Expect(bi.String()).To(Equal(data))
		actual := new(big.Int)
		actual.SetString(data, 10)
		Expect(actual).To(Equal(bi))
	})

	It("throws error when can't connect to the database", func() {
		invalidDatabase := &postgres.Config{}
		node := node.Info{GenesisBlock: "GENESIS", NetworkID: "1", ID: "x123", ClientName: "geth"}

		_, err := postgres.NewDB(invalidDatabase, node, true)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(postgres.DbConnectionFailedMsg))
	})

	It("throws error when can't create node", func() {
		badHash := fmt.Sprintf("x %s", strings.Repeat("1", 100))
		node := node.Info{GenesisBlock: badHash, NetworkID: "1", ID: "x123", ClientName: "geth"}

		_, err := shared.SetupDBWithNode(node)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(postgres.SettingNodeFailedMsg))
	})
})
