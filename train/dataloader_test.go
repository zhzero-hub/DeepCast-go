package train

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
)

var ctx = context.Background()

func TestReadFromFile(t *testing.T) {
	pwd, _ := os.Getwd()
	csvData, _ := ReadFromFile(pwd+"/../data/", "user_viewing_dataset.csv", "csv")
	countMap := make(map[string]int)
	for _, data := range csvData {
		if count, ok := countMap[data[0]]; ok {
			countMap[data[0]] = count + 1
		} else {
			countMap[data[0]] = 1
		}
	}
}

func TestReadFromFile2(t *testing.T) {
	pwd, _ := os.Getwd()
	data, _ := ReadFromFile(pwd+"/../data/wsdm_train_data/", "video_related_data.csv", "csv")
	fmt.Println(len(data))
}

func TestReadFromFile3(t *testing.T) {
	dataset, err := LoadUserViewingDataset(ctx)
	if err != nil {
		return
	}
	fmt.Println(dataset)
}

func TestReadFromFile4(t *testing.T) {
	liverMap := make(map[string]int64, 0)
	data, _ := ReadFromFile("../data/", "user_viewing_dataset.csv", "csv")
	for _, csvData := range data {
		if count, ok := liverMap[csvData[2]]; !ok {
			liverMap[csvData[2]] = 1
		} else {
			liverMap[csvData[2]] = count + 1
		}
	}
	var topCount []struct {
		Liver string
		Count int64
	}
	for liver, count := range liverMap {
		topCount = append(topCount, struct {
			Liver string
			Count int64
		}{Liver: liver, Count: count})
	}
	sort.Slice(topCount, func(i, j int) bool {
		return topCount[i].Count > topCount[j].Count
	})
	fmt.Println(liverMap)
}
