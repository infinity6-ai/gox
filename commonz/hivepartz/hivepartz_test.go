package hivepartz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/util/filez/hivepartz"
)

func TestUnitHiveParts(t *testing.T) {

	assert.PanicsWithValue(t, `error validation fail True, params: {"actual":false,"msg":"path is not relative: /a/b"}`, func() {
		hivepartz.Parse("/a/b", -1)
	})

	assert.PanicsWithValue(t, `error validation fail True, params: {"actual":false,"msg":"path is not relative: /"}`, func() {
		hivepartz.Parse("/", -1)
	})

	assertParse(t, &hivepartz.HiveParts{}, "", "", -1)

	assertParse(t, &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
		{Name: "c", Value: "3"},
	}}, "r1/r2.txt", "a=1/b=2/c=3/r1/r2.txt", -1)

	assertParse(t, &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
		{Name: "c", Value: "3"},
	}}, "", "a=1/b=2/c=3/", -1)

	assertParse(t, &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
		{Name: "c", Value: "3"},
	}}, "", "a=1/b=2/c=3", -1)

	assertParse(t, &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
		{Name: "c", Value: "3"},
	}}, "r1/r2.txt", "a=1/b=2/c=3/r1/r2.txt", 3)

	assertParse(t, &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
	}}, "c=3/r1/r2.txt", "a=1/b=2/c=3/r1/r2.txt", 2)

	assertParse(t, &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
	}}, "b=2/c=3/r1/r2.txt", "a=1/b=2/c=3/r1/r2.txt", 1)

	assertParse(t, &hivepartz.HiveParts{}, "a=1/b=2/c=3/r1/r2.txt", "a=1/b=2/c=3/r1/r2.txt", 0)

}

func TestUnitToMap(t *testing.T) {
	hp := &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
	}}
	assert.Equal(t, map[string]string{"a": "1", "b": "2"}, hp.ToMap())
}

func TestUnitFormat(t *testing.T) {
	hp := &hivepartz.HiveParts{Parts: []*hivepartz.HivePart{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
	}}
	assert.Equal(t, "a=1/b=2", hp.Format())
}

func assertParse(t *testing.T, expectedHiveParts *hivepartz.HiveParts, expectedRemaining, f string, max int) {
	hp, r := hivepartz.Parse(f, max)
	assert.Equal(t, expectedHiveParts, hp)
	assert.Equal(t, expectedRemaining, r)
}
