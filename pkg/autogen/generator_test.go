// VulcanizeDB
// Copyright © 2018 Vulcanize

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

package autogen_test

import (
	"plugin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/vulcanize/vulcanizedb/libraries/shared/transformer"
	"github.com/vulcanize/vulcanizedb/libraries/shared/watcher"
	"github.com/vulcanize/vulcanizedb/pkg/autogen"
	"github.com/vulcanize/vulcanizedb/pkg/autogen/test_helpers"
	"github.com/vulcanize/vulcanizedb/pkg/config"
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres/repositories"
	"github.com/vulcanize/vulcanizedb/pkg/transformers/bite"
	"github.com/vulcanize/vulcanizedb/utils"
)

var localConfig = autogen.Config{
	Initializers: map[string]string{
		"bite": "github.com/vulcanize/vulcanizedb/pkg/autogen/test_helpers/bite",
		"deal": "github.com/vulcanize/vulcanizedb/pkg/autogen/test_helpers/deal",
	},
	FileName: "localTestTransformerSet",
	FilePath: "$GOPATH/src/github.com/vulcanize/vulcanizedb/pkg/autogen/test_helpers/test/",
}

var externalConfig = autogen.Config{
	Initializers: map[string]string{
		"bite": "github.com/vulcanize/mcd_transformers/transformers/bite",
		"deal": "github.com/vulcanize/mcd_transformers/transformers/deal",
	},
	Dependencies: map[string]string{
		"mcd_transformers": "github.com/vulcanize/mcd_transformers",
	},
	FileName: "externalTestTransformerSet",
	FilePath: "$GOPATH/src/github.com/vulcanize/vulcanizedb/pkg/autogen/test_helpers/test/",
}

type Exporter interface {
	Export() []transformer.TransformerInitializer
}

var _ = Describe("Generator test", func() {
	var g autogen.Generator
	var goPath, soPath string
	var err error
	var bc core.BlockChain
	var db *postgres.DB
	var hr repositories.HeaderRepository
	var headerID int64
	viper.SetConfigName("compose")
	viper.AddConfigPath("$GOPATH/src/github.com/vulcanize/vulcanizedb/environments/")

	Describe("Using local config", func() {
		BeforeEach(func() {
			goPath, soPath, err = localConfig.GetPluginPaths()
			Expect(err).ToNot(HaveOccurred())
			g = autogen.NewGenerator(localConfig, config.Database{})
			err = g.GenerateExporterPlugin()
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := utils.ClearFiles(goPath, soPath)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("GenerateTransformerPlugin", func() {
			It("It bundles the specified transformer initializers into a Exporter object and creates .so", func() {
				plug, err := plugin.Open(soPath)
				Expect(err).ToNot(HaveOccurred())
				symExporter, err := plug.Lookup("Exporter")
				Expect(err).ToNot(HaveOccurred())
				exporter, ok := symExporter.(Exporter)
				Expect(ok).To(Equal(true))
				initializers := exporter.Export()
				Expect(len(initializers)).To(Equal(2))
			})

			It("Loads our generated Exporter and uses it to import an arbitrary set of TransformerInitializers that we can execute over", func() {
				db, bc = test_helpers.SetupDBandBC()
				defer test_helpers.TearDown(db)

				hr = repositories.NewHeaderRepository(db)
				header1, err := bc.GetHeaderByNumber(9377319)
				Expect(err).ToNot(HaveOccurred())
				headerID, err = hr.CreateOrUpdateHeader(header1)
				Expect(err).ToNot(HaveOccurred())

				plug, err := plugin.Open(soPath)
				Expect(err).ToNot(HaveOccurred())
				symExporter, err := plug.Lookup("Exporter")
				Expect(err).ToNot(HaveOccurred())
				exporter, ok := symExporter.(Exporter)
				Expect(ok).To(Equal(true))
				initializers := exporter.Export()

				w := watcher.NewWatcher(db, bc)
				w.AddTransformers(initializers)
				err = w.Execute()
				Expect(err).ToNot(HaveOccurred())

				type model struct {
					bite.BiteModel
					Id       int64 `db:"id"`
					HeaderId int64 `db:"header_id"`
				}

				returned := model{}

				err = db.Get(&returned, `SELECT * FROM maker.bite WHERE header_id = $1`, headerID)
				Expect(err).ToNot(HaveOccurred())
				Expect(returned.Ilk).To(Equal("ETH"))
				Expect(returned.Urn).To(Equal("0x0000d8b4147eDa80Fec7122AE16DA2479Cbd7ffB"))
				Expect(returned.Ink).To(Equal("80000000000000000000"))
				Expect(returned.Art).To(Equal("11000000000000000000000"))
				Expect(returned.IArt).To(Equal("12496609999999999999992"))
				Expect(returned.Tab).To(Equal("11000000000000000000000"))
				Expect(returned.NFlip).To(Equal("7"))
				Expect(returned.TransactionIndex).To(Equal(uint(1)))
				Expect(returned.LogIndex).To(Equal(uint(4)))
			})
		})
	})

	Describe("Using external config", func() {
		BeforeEach(func() {
			goPath, soPath, err = externalConfig.GetPluginPaths()
			Expect(err).ToNot(HaveOccurred())
			g = autogen.NewGenerator(externalConfig, config.Database{})
			err = g.GenerateExporterPlugin()
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := utils.ClearFiles(goPath, soPath)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("GenerateTransformerPlugin", func() {
			It("It bundles the specified transformer initializers into a Exporter object and creates .so", func() {
				plug, err := plugin.Open(soPath)
				Expect(err).ToNot(HaveOccurred())
				symExporter, err := plug.Lookup("Exporter")
				Expect(err).ToNot(HaveOccurred())
				exporter, ok := symExporter.(Exporter)
				Expect(ok).To(Equal(true))
				initializers := exporter.Export()
				Expect(len(initializers)).To(Equal(2))
			})

			It("Loads our generated Exporter and uses it to import an arbitrary set of TransformerInitializers that we can execute over", func() {
				db, bc = test_helpers.SetupDBandBC()
				defer test_helpers.TearDown(db)

				hr = repositories.NewHeaderRepository(db)
				header1, err := bc.GetHeaderByNumber(9377319)
				Expect(err).ToNot(HaveOccurred())
				headerID, err = hr.CreateOrUpdateHeader(header1)
				Expect(err).ToNot(HaveOccurred())

				plug, err := plugin.Open(soPath)
				Expect(err).ToNot(HaveOccurred())
				symExporter, err := plug.Lookup("Exporter")
				Expect(err).ToNot(HaveOccurred())
				exporter, ok := symExporter.(Exporter)
				Expect(ok).To(Equal(true))
				initializers := exporter.Export()

				w := watcher.NewWatcher(db, bc)
				w.AddTransformers(initializers)
				err = w.Execute()
				Expect(err).ToNot(HaveOccurred())

				type model struct {
					bite.BiteModel
					Id       int64 `db:"id"`
					HeaderId int64 `db:"header_id"`
				}

				returned := model{}

				err = db.Get(&returned, `SELECT * FROM maker.bite WHERE header_id = $1`, headerID)
				Expect(err).ToNot(HaveOccurred())
				Expect(returned.Ilk).To(Equal("ETH"))
				Expect(returned.Urn).To(Equal("0x0000d8b4147eDa80Fec7122AE16DA2479Cbd7ffB"))
				Expect(returned.Ink).To(Equal("80000000000000000000"))
				Expect(returned.Art).To(Equal("11000000000000000000000"))
				Expect(returned.IArt).To(Equal("12496609999999999999992"))
				Expect(returned.Tab).To(Equal("11000000000000000000000"))
				Expect(returned.NFlip).To(Equal("7"))
				Expect(returned.TransactionIndex).To(Equal(uint(1)))
				Expect(returned.LogIndex).To(Equal(uint(4)))
			})
		})
	})
})
