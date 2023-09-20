package interop_test

import (
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
