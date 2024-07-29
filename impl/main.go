package main

import (
	"fmt"

	"github.com/Andrey-Dubas/image-server/image_repository"
)

func testImageRepo() {
	var repo image_repository.IImageRepository
	repo, err := image_repository.NewPostgresImageRepository("localhost", "1", "postgres", "postgres")
	if err != nil {
		fmt.Printf("%+x", err)
	}

	repoStats, err := repo.GetStatistics()
	if err != nil {
		fmt.Printf("%+x", err)
	}

	fmt.Printf("statistics: %+v", repoStats)
}

func main() {
	testImageRepo()
}
