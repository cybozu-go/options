package interop_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/cybozu-go/options"
)

type Row struct {
	ID   int64
	Num  options.Option[int64]
	Str  options.Option[string]
	Ts   options.Option[time.Time]
	Blob options.Option[[]byte]
}

func TestGoCmp(t *testing.T) {
	row1 := &Row{
		ID:   1,
		Num:  options.None[int64](),
		Str:  options.None[string](),
		Ts:   options.None[time.Time](),
		Blob: options.None[[]byte](),
	}
	row2 := &Row{
		ID:   1,
		Num:  options.New(int64(3)),
		Str:  options.New("hello"),
		Ts:   options.New(time.Now().UTC().Round(time.Microsecond)),
		Blob: options.New([]byte("world")),
	}

	shouldEqual := func(a, b *Row) {
		t.Helper()
		if diff := cmp.Diff(a, b); diff != "" {
			t.Errorf("should be equal, but not:\n%s", diff)
		}
	}
	shouldNotEqual := func(a, b *Row) {
		t.Helper()
		if diff := cmp.Diff(a, b); diff == "" {
			t.Errorf("should have diff, but no diff found")
		}
	}

	shouldEqual(row1, row1)
	shouldEqual(row2, row2)
	shouldNotEqual(row1, row2)
	shouldNotEqual(row2, row1)
}

var (
	reNonprintable = regexp.MustCompile(`[^[:print:]]+`)
	reSpaces       = regexp.MustCompile(`[[:space:]]+`)
)

func equalsIgnoringSpaces(a, b string) bool {
	a = reNonprintable.ReplaceAllString(a, "")
	a = reSpaces.ReplaceAllString(a, "")
	a = strings.TrimSpace(a)
	b = reNonprintable.ReplaceAllString(b, "")
	b = reSpaces.ReplaceAllString(b, "")
	b = strings.TrimSpace(b)
	return a == b
}

type NestedData struct {
	Value string
}

type TestData struct {
	Value  string
	Nested *NestedData
}

func TestGoCmp_Transformer(t *testing.T) {
	cmpopt := cmp.Transformer("options.Option", options.Pointer[*TestData])

	d1 := options.New(&TestData{
		Value: "test",
		Nested: &NestedData{
			Value: "test",
		},
	})
	d2 := options.New(&TestData{
		Value: "test",
		Nested: &NestedData{
			Value: "test2",
		},
	})

	expectedDiff := `
		options.Option[*github.com/cybozu-go/options/interop_test.TestData](Inverse(options.Option, &&interop_test.TestData{
			Value:  "test",
	  - 	Nested: &interop_test.NestedData{Value: "test"},
	  + 	Nested: &interop_test.NestedData{Value: "test2"},
		}))
	`
	actualDiff := cmp.Diff(d1, d2, cmpopt)
	if !equalsIgnoringSpaces(actualDiff, expectedDiff) {
		t.Errorf("unexpected diff.\n[expected]\n%s\n\n[actual]\n%s", expectedDiff, actualDiff)
	}
}

func TestSQL(t *testing.T) {
	testCases := []struct {
		title    string
		inserted *Row
	}{
		{
			title: "Present",
			inserted: &Row{
				ID:   1,
				Num:  options.New(int64(3)),
				Str:  options.New("hello"),
				Ts:   options.New(time.Now().UTC().Round(time.Microsecond)),
				Blob: options.New([]byte("world")),
			},
		},
		{
			title: "None",
			inserted: &Row{
				ID:   1,
				Num:  options.None[int64](),
				Str:  options.None[string](),
				Ts:   options.None[time.Time](),
				Blob: options.None[[]byte](),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			db, err := sqlx.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })

			// CREATE TABLE
			if _, err := db.Exec("" +
				"CREATE TABLE `test` (" +
				"  `id` INTEGER PRIMARY KEY, " +
				"  `num` INTEGER, " +
				"  `str` TEXT, " +
				"  `ts` DATETIME, " +
				"  `blob` BLOB " +
				")",
			); err != nil {
				t.Fatal(err)
			}

			// INSERT
			if _, err := db.NamedExec(
				"INSERT INTO `test` VALUES (:id, :num, :str, :ts, :blob)",
				tc.inserted,
			); err != nil {
				t.Fatal(err)
			}

			// SELECT
			var selected Row
			if err := db.Get(
				&selected,
				"SELECT * FROM `test` WHERE `id` = ?",
				tc.inserted.ID,
			); err != nil {
				t.Fatal(err)
			}

			// Verify
			if diff := cmp.Diff(tc.inserted, &selected); diff != "" {
				t.Errorf("row mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
