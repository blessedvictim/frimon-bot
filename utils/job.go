package utils

import (
	"github.com/blessedvictim/frimon-bot/model"
	"math/rand"
	"sync"
)

func GetCurrentJob(job *model.Job, ch chan string) *model.Content {
	var wg sync.WaitGroup
	var currentContentList []model.Content
	wg.Add(1)
	go func(ch chan string) {
		defer wg.Done()
		select {
		case lastContent := <-ch:
			for i, _ := range job.ContentList {
				if job.ContentList[i].ID != lastContent {
					currentContentList = append(currentContentList, job.ContentList[i])
				}
			}
		default:
			currentContentList = job.ContentList
		}
	}(ch)
	wg.Wait()
	i := rand.Intn(len(currentContentList))
	content := currentContentList[i]
	ch <- content.ID
	return &content
}
