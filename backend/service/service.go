package service

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"main/model"
)

// Service
type Service struct {
	csvr   *os.File
	csvw   *csv.Writer
	dbPath string
}

// New creates a new Service layer
func New(
	csvr *os.File,
	csvw *csv.Writer,
	dbPath string) (*Service, error) {

	return &Service{csvr, csvw, dbPath}, nil
}

// GetMovies -
func (s *Service) GetMovies() ([]*model.Movie, error) {
	out := []*model.Movie{}

	csvr := csv.NewReader(s.csvr)

	data, err := csvr.ReadAll()
	if err != nil {
		fmt.Print("Error reading")
		return nil, err
	}

	for _, row := range data {
		movie := model.Movie{
			ImdbTitleId:         row[0],
			Title:               row[1],
			OriginalTitle:       row[2],
			Year:                row[3],
			DatePublished:       row[4],
			Genre:               row[5],
			Duration:            row[6],
			Country:             row[7],
			Language:            row[8],
			Director:            row[9],
			Writer:              row[10],
			ProductionCompany:   row[11],
			Actors:              row[12],
			Description:         row[13],
			AvgVote:             row[14],
			Votes:               row[15],
			Budget:              row[16],
			UsaGrossIncome:      row[17],
			WorlwideGrossIncome: row[18],
			Metascore:           row[19],
			ReviewsFromUsers:    row[20],
			ReviewsFromCritics:  row[21],
			Poster:              row[22],
		}
		out = append(out, &movie)
	}
	s.csvr.Seek(0, 0)

	return out, nil
}

// GetMovieById -
func (s *Service) GetMovieById(movieID string) (*model.Movie, error) {
	out := &model.Movie{}

	csvr := csv.NewReader(s.csvr)

	data, err := csvr.ReadAll()
	if err != nil {
		fmt.Print("Error reading")
		return nil, err
	}

	for _, row := range data {
		if row[0] == movieID {
			out.ImdbTitleId = row[0]
			out.Title = row[1]
			out.Year = row[2]
		}
	}
	s.csvr.Seek(0, 0)

	return out, nil
}

// GetMovieById -
func (s *Service) GetConcurrently(queryParams model.QueryParameters, complete bool, id string) ([]interface{}, error) {

	file, err := os.Open(s.dbPath)
	if err != nil {
		// requestErrors = append(requestErrors, err.Error())
		log.Fatal(err)
	}
	defer file.Close()

	out := []interface{}{}

	jobs := make(chan []string)
	results := make(chan interface{})

	wg := new(sync.WaitGroup)

	// start workers
	var workers int = 1
	switch {
	case queryParams.Items <= 50:
		workers = 2
	case queryParams.Items > 50 && queryParams.Items < 500:
		workers = 25
	case queryParams.Items >= 500:
		workers = 100
	default:
		workers = 1
	}

	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go worker(jobs, results, wg, queryParams, complete, id)
	}

	// scan the file into the string channel
	go func() {

		csvr := csv.NewReader(file)
		data, err := csvr.ReadAll()
		if err != nil {
			fmt.Print("Error reading")
			close(jobs)
		}

		for each := range data {
			job := data[each]
			jobs <- job
		}
		close(jobs)
	}()

	// Collect all the results,  make sure we close the result channel when everything was processed
	go func() {
		wg.Wait()
		close(results)
	}()

	// Convert channel to slice of Movie and send
	movieCounter := 0
	for movieInterface := range results {
		if movieCounter == queryParams.Items {
			break
		}
		out = append(out, movieInterface)
		movieCounter++
	}
	log.Println("service -> GetConcurrently ", len(out))
	return out, nil
}

func worker(jobs <-chan []string, results chan<- interface{}, wg *sync.WaitGroup, queryParams model.QueryParameters, complete bool, id string) {
	defer wg.Done()

	for lineItems := range jobs {
		if complete && id != "" && id == lineItems[0] {
			log.Println("Valuating the Movie", complete, id)
			movie := model.Movie{
				ImdbTitleId:         lineItems[0],
				Title:               lineItems[1],
				OriginalTitle:       lineItems[2],
				Year:                lineItems[3],
				DatePublished:       lineItems[4],
				Genre:               lineItems[5],
				Duration:            lineItems[6],
				Country:             lineItems[7],
				Language:            lineItems[8],
				Director:            lineItems[9],
				Writer:              lineItems[10],
				ProductionCompany:   lineItems[11],
				Actors:              lineItems[12],
				Description:         lineItems[13],
				AvgVote:             lineItems[14],
				Votes:               lineItems[15],
				Budget:              lineItems[16],
				UsaGrossIncome:      lineItems[17],
				WorlwideGrossIncome: lineItems[18],
				Metascore:           lineItems[19],
				ReviewsFromUsers:    lineItems[20],
				ReviewsFromCritics:  lineItems[21],
				Poster:              lineItems[22],
			}
			results <- movie
		}
		if !complete {
			// get id from Movie struct and parse the string to a number
			idOfCurrentMovie := lineItems[0]            // get id of current movie
			substringOfId := idOfCurrentMovie[2:]       // convert to only string numbers
			integerId, _ := strconv.Atoi(substringOfId) // parse substring to int

			log.Println("Valuating the MovieSummary", complete, id, idOfCurrentMovie)

			// if numberType is supposed to be odd and it is not, then continue to next line wihtout adding it to the list
			if queryParams.Type == "odd" && !Odd(integerId) {
				continue
			}
			// if numberType is supposed to be even and it is not, then continue to next line wihtout adding it to the list
			if queryParams.Type == "even" && !Even(integerId) {
				continue
			}

			// if it got to this point add it to the list
			movieSummary := model.MovieSummary{
				ImdbTitleId:   lineItems[0],
				OriginalTitle: lineItems[2],
				Year:          lineItems[3],
				Poster:        lineItems[22],
			}
			results <- movieSummary

		}
	}
}

func Even(number int) bool {
	return number%2 == 0
}

func Odd(number int) bool {
	return !Even(number)
}
