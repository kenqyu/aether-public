// Backend > ResponseGenerator
// This file provides a set of functions that take a database response, and convert it into a set of paginated (or nonpaginated) results.

package responsegenerator

import (
	// "fmt"
	"aether-core/io/api"
	"aether-core/io/persistence"
	"aether-core/services/globals"
	"aether-core/services/logging"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// GeneratePrefilledApiResponse constructs the basic ApiResponse and fills it with the data about the local machine.
func GeneratePrefilledApiResponse() *api.ApiResponse {
	var resp api.ApiResponse
	resp.NodeId = api.Fingerprint(globals.NodeId)
	resp.Address.LocationType = uint8(globals.AddressType)
	resp.Address.Port = uint16(globals.AddressPort)
	resp.Address.Protocol.VersionMajor = uint8(globals.ProtocolVersionMajor)
	resp.Address.Protocol.VersionMinor = uint16(globals.ProtocolVersionMinor)
	resp.Address.Protocol.Extensions = globals.ProtocolExtensions
	resp.Address.Client.VersionMajor = uint8(globals.ClientVersionMajor)
	resp.Address.Client.VersionMinor = uint16(globals.ClientVersionMinor)
	resp.Address.Client.VersionPatch = uint16(globals.ClientVersionPatch)
	resp.Address.Client.ClientName = globals.ClientName
	return &resp
}

func ConvertApiResponseToJson(resp *api.ApiResponse) ([]byte, error) {
	result, err := json.Marshal(resp)
	var jsonErr error
	if err != nil {
		jsonErr = errors.New(fmt.Sprint(
			"This ApiResponse failed to convert to JSON. Error: %#v, ApiResponse: %#v", err, *resp))
	}
	return result, jsonErr
}

type FilterSet struct {
	Fingerprints []api.Fingerprint
	TimeStart    api.Timestamp
	TimeEnd      api.Timestamp
	Embeds       []string
}

func processFilters(req *api.ApiResponse) FilterSet {
	var fs FilterSet
	for _, filter := range req.Filters {
		// Fingerprint
		if filter.Type == "fingerprint" {
			for _, fp := range filter.Values {
				fs.Fingerprints = append(fs.Fingerprints, api.Fingerprint(fp))
			}
		}
		// Embeds
		if filter.Type == "embed" {
			for _, embed := range filter.Values {
				fs.Embeds = append(fs.Embeds, embed)
			}
		}
		// If a time filter is given, timeStart is either the timestamp provided by the remote if it's larger than the end date of the last cache, or the end timestamp of the last cache.
		// In essence, we do not provide anything that is already cached from the live server.
		if filter.Type == "timestamp" {
			// now := int64(time.Now().Unix())
			start, _ := strconv.ParseInt(filter.Values[0], 10, 64)
			end, _ := strconv.ParseInt(filter.Values[1], 10, 64)

			// If there is a value given (not 0), that is, the timerange filter is active.
			// The sanitisation of these ranges are done in the DB level, so this is just intake.
			if start > 0 || end > 0 {
				fs.TimeStart = api.Timestamp(start)
				fs.TimeEnd = api.Timestamp(end)
			}

		}
	}
	return fs
}

func splitEntityIndexesToPages(fullData *api.Response) *[]api.Response {
	var entityTypes []string
	if len(fullData.BoardIndexes) > 0 {
		entityTypes = append(entityTypes, "boardindexes")
	}
	if len(fullData.ThreadIndexes) > 0 {
		entityTypes = append(entityTypes, "threadindexes")
	}
	if len(fullData.PostIndexes) > 0 {
		entityTypes = append(entityTypes, "postindexes")
	}
	if len(fullData.VoteIndexes) > 0 {
		entityTypes = append(entityTypes, "voteindexes")
	}
	if len(fullData.AddressIndexes) > 0 {
		entityTypes = append(entityTypes, "addressindexes")
	}
	if len(fullData.KeyIndexes) > 0 {
		entityTypes = append(entityTypes, "keyindexes")
	}
	if len(fullData.TruststateIndexes) > 0 {
		entityTypes = append(entityTypes, "truststateindexes")
	}

	var pages []api.Response
	// This is a lot of copy paste. This is because there is no automatic conversion from []api.Boards being recognised as []api.Provable. Without that, I have to convert them explicitly to be able to put them into a map[string:struct] which is a lot of extra work - more work than copy paste.
	for i, _ := range entityTypes {
		// Index entities
		if entityTypes[i] == "boardindexes" {
			dataSet := fullData.BoardIndexes
			pageSize := globals.EntityPageSizesObj.BoardIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.BoardIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "threadindexes" {
			dataSet := fullData.ThreadIndexes
			pageSize := globals.EntityPageSizesObj.ThreadIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.ThreadIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "postindexes" {
			dataSet := fullData.PostIndexes
			pageSize := globals.EntityPageSizesObj.PostIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.PostIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "voteindexes" {
			dataSet := fullData.VoteIndexes
			pageSize := globals.EntityPageSizesObj.VoteIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.VoteIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "keyindexes" {
			dataSet := fullData.KeyIndexes
			pageSize := globals.EntityPageSizesObj.KeyIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.KeyIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "addressindexes" {
			dataSet := fullData.AddressIndexes
			pageSize := globals.EntityPageSizesObj.AddressIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.AddressIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "truststateindexes" {
			dataSet := fullData.TruststateIndexes
			pageSize := globals.EntityPageSizesObj.TruststateIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.TruststateIndexes = pageData
				pages = append(pages, page)
			}
		}
	}
	if len(entityTypes) == 0 {
		// The result is empty
		var page api.Response
		pages = append(pages, page)
	}
	return &pages
}

func splitEntitiesToPages(fullData *api.Response) *[]api.Response {
	var entityTypes []string
	// We do this check set below so that we don't run pagination logic on entity types that does not exist in this response. This is a bit awkward because there's no good way to iterate over fields of a struct.
	if len(fullData.Boards) > 0 {
		entityTypes = append(entityTypes, "boards")
	}
	if len(fullData.BoardIndexes) > 0 {
		entityTypes = append(entityTypes, "boardindexes")
	}
	if len(fullData.Threads) > 0 {
		entityTypes = append(entityTypes, "threads")
	}
	if len(fullData.ThreadIndexes) > 0 {
		entityTypes = append(entityTypes, "threadindexes")
	}
	if len(fullData.Posts) > 0 {
		entityTypes = append(entityTypes, "posts")
	}
	if len(fullData.PostIndexes) > 0 {
		entityTypes = append(entityTypes, "postindexes")
	}
	if len(fullData.Votes) > 0 {
		entityTypes = append(entityTypes, "votes")
	}
	if len(fullData.VoteIndexes) > 0 {
		entityTypes = append(entityTypes, "voteindexes")
	}
	if len(fullData.Addresses) > 0 {
		entityTypes = append(entityTypes, "addresses")
	}
	if len(fullData.AddressIndexes) > 0 {
		entityTypes = append(entityTypes, "addressindexes")
	}
	if len(fullData.Keys) > 0 {
		entityTypes = append(entityTypes, "keys")
	}
	if len(fullData.KeyIndexes) > 0 {
		entityTypes = append(entityTypes, "keyindexes")
	}
	if len(fullData.Truststates) > 0 {
		entityTypes = append(entityTypes, "truststates")
	}
	if len(fullData.TruststateIndexes) > 0 {
		entityTypes = append(entityTypes, "truststateindexes")
	}

	var pages []api.Response
	// This is a lot of copy paste. This is because there is no automatic conversion from []api.Boards being recognised as []api.Provable. Without that, I have to convert them explicitly to be able to put them into a map[string:struct] which is a lot of extra work - more work than copy paste.
	for i, _ := range entityTypes {
		if entityTypes[i] == "boards" {
			dataSet := fullData.Boards
			pageSize := globals.EntityPageSizesObj.Boards
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Boards = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "threads" {
			dataSet := fullData.Threads
			pageSize := globals.EntityPageSizesObj.Threads
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Threads = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "posts" {
			dataSet := fullData.Posts
			pageSize := globals.EntityPageSizesObj.Posts
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Posts = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "votes" {
			dataSet := fullData.Votes
			pageSize := globals.EntityPageSizesObj.Votes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Votes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "addresses" {
			dataSet := fullData.Addresses
			pageSize := globals.EntityPageSizesObj.Addresses
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Addresses = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "keys" {
			dataSet := fullData.Keys
			pageSize := globals.EntityPageSizesObj.Keys
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Keys = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "truststates" {
			dataSet := fullData.Truststates
			pageSize := globals.EntityPageSizesObj.Truststates
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.Truststates = pageData
				pages = append(pages, page)
			}
		}
		// Index entities
		if entityTypes[i] == "boardindexes" {
			dataSet := fullData.BoardIndexes
			pageSize := globals.EntityPageSizesObj.BoardIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.BoardIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "threadindexes" {
			dataSet := fullData.ThreadIndexes
			pageSize := globals.EntityPageSizesObj.ThreadIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.ThreadIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "postindexes" {
			dataSet := fullData.PostIndexes
			pageSize := globals.EntityPageSizesObj.PostIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.PostIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "voteindexes" {
			dataSet := fullData.VoteIndexes
			pageSize := globals.EntityPageSizesObj.VoteIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.VoteIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "keyindexes" {
			dataSet := fullData.KeyIndexes
			pageSize := globals.EntityPageSizesObj.KeyIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.KeyIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "addressindexes" {
			dataSet := fullData.AddressIndexes
			pageSize := globals.EntityPageSizesObj.AddressIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.AddressIndexes = pageData
				pages = append(pages, page)
			}
		}
		if entityTypes[i] == "truststateindexes" {
			dataSet := fullData.TruststateIndexes
			pageSize := globals.EntityPageSizesObj.TruststateIndexes
			numPages := len(dataSet)/pageSize + 1
			// The division above is floored.
			for i := 0; i < numPages; i++ {
				beg := i * pageSize
				var end int
				// This is to protect from 'slice bounds out of range'
				if (i+1)*pageSize > len(dataSet) {
					end = len(dataSet)
				} else {
					end = (i + 1) * pageSize
				}
				pageData := dataSet[beg:end]
				var page api.Response
				page.TruststateIndexes = pageData
				pages = append(pages, page)
			}
		}
	}
	if len(entityTypes) == 0 {
		// The result is empty
		var page api.Response
		pages = append(pages, page)
	}
	return &pages
}

func convertResponsesToApiResponses(r *[]api.Response) *[]api.ApiResponse {
	var responses []api.ApiResponse
	for i, _ := range *r {
		resp := GeneratePrefilledApiResponse()
		resp.ResponseBody.Boards = (*r)[i].Boards
		resp.ResponseBody.Threads = (*r)[i].Threads
		resp.ResponseBody.Posts = (*r)[i].Posts
		resp.ResponseBody.Votes = (*r)[i].Votes
		resp.ResponseBody.Addresses = (*r)[i].Addresses
		resp.ResponseBody.Keys = (*r)[i].Keys
		resp.ResponseBody.Truststates = (*r)[i].Truststates
		// Indexes
		resp.ResponseBody.BoardIndexes = (*r)[i].BoardIndexes
		resp.ResponseBody.ThreadIndexes = (*r)[i].ThreadIndexes
		resp.ResponseBody.PostIndexes = (*r)[i].PostIndexes
		resp.ResponseBody.VoteIndexes = (*r)[i].VoteIndexes
		resp.ResponseBody.AddressIndexes = (*r)[i].AddressIndexes
		resp.ResponseBody.KeyIndexes = (*r)[i].KeyIndexes
		resp.ResponseBody.TruststateIndexes = (*r)[i].TruststateIndexes
		resp.Pagination.Pages = uint64(len(*r) - 1) // pagination starts from 0
		resp.Pagination.CurrentPage = uint64(i)
		responses = append(responses, *resp)
	}
	return &responses
}

func generateRandomHash() (string, error) {
	const LETTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	saltBytes := make([]byte, 16)
	for i := range saltBytes {
		randNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(LETTERS))))
		if err != nil {
			return "", errors.New(fmt.Sprint(
				"Random number generator generated an error. err: ", err))
		}
		saltBytes[i] = LETTERS[int(randNum.Int64())]
	}
	calculator := sha256.New()
	calculator.Write(saltBytes)
	resultHex := fmt.Sprintf("%x", calculator.Sum(nil))
	return resultHex, nil
}

func generateExpiryTimestamp() int64 {
	expiry := time.Duration(globals.PostResponseExpiryMinutes) * time.Minute
	expiryTs := int64(time.Now().Add(expiry).Unix())
	return expiryTs
}

func findEntityInApiResponse(resp api.ApiResponse) string {
	if len(resp.ResponseBody.Boards) > 0 {
		return "boards"
	}
	if len(resp.ResponseBody.Threads) > 0 {
		return "threads"
	}
	if len(resp.ResponseBody.Posts) > 0 {
		return "posts"
	}
	if len(resp.ResponseBody.Votes) > 0 {
		return "votes"
	}
	if len(resp.ResponseBody.Addresses) > 0 {
		return "addresses"
	}
	if len(resp.ResponseBody.Keys) > 0 {
		return "keys"
	}
	if len(resp.ResponseBody.Truststates) > 0 {
		return "truststates"
	}
	return ""
}

func createPath(path string) {
	os.MkdirAll(path, 0755)
}

func saveFileToDisk(fileContents []byte, path string, filename string) {
	ioutil.WriteFile(fmt.Sprint(path, "/", filename), fileContents, 0755)
}

// bakeFinalApiResponse looks at the resultpages. If there is one, it is directly provided as is. If there is more, the results are committed into the file system, and a cachelink page is provided instead.
func bakeFinalApiResponse(resultPages *[]api.ApiResponse) (*api.ApiResponse, error) {
	resp := GeneratePrefilledApiResponse()
	if len(*resultPages) > 1 {
		// Create a random SHA256 hash as folder name
		dirname, err := generateRandomHash()
		if err != nil {
			return resp, err
		}
		// Generate the responses directory if doesn't exist. Add the expiry date to the folder name to be searched for.
		foldername := fmt.Sprint(generateExpiryTimestamp(), "_", dirname)
		responsedir := fmt.Sprint(globals.UserDirectory, "/statics/responses/", foldername)
		os.MkdirAll(responsedir, 0755)
		var jsons [][]byte
		// For each response, number it, set timestamps etc. And save to disk.
		for i, _ := range *resultPages {
			resultPage := (*resultPages)[i]
			entityType := findEntityInApiResponse(resultPage)
			// Set timestamp, number of items in it, total page count, and which page.

			resultPage.Pagination.Pages = uint64(len(*resultPages))
			resultPage.Pagination.CurrentPage = uint64(i)
			resultPage.Timestamp = api.Timestamp(time.Now().Unix())
			resultPage.Entity = entityType
			resultPage.Endpoint = fmt.Sprint(entityType, "_post")
			jsonResp, err := ConvertApiResponseToJson(&resultPage)
			if err != nil {
				logging.Log(1, fmt.Sprintf("This page of a multiple-page post response failed to convert to JSON. Error: %#v\n, Request Body: %#v\n", err, resultPage))
			}
			jsons = append(jsons, jsonResp)
		}
		// Insert these jsons into the filesystem.
		for i, _ := range jsons {
			name := fmt.Sprint(i, ".json")
			createPath(responsedir)
			saveFileToDisk(jsons[i], responsedir, name)
			var c api.ResultCache
			c.ResponseUrl = foldername
			resp.Results = append(resp.Results, c)
		}
		resp.Endpoint = "multipart_post_response"

	} else if len(*resultPages) == 1 {
		// There is only one response page here.
		entityType := findEntityInApiResponse((*resultPages)[0])
		resp.Pagination.Pages = 0 // These start to count from 0
		resp.Pagination.CurrentPage = 0
		resp.Entity = entityType
		resp.Endpoint = "singular_post_response"
		resp.ResponseBody = (*resultPages)[0].ResponseBody
	} else {
		logging.LogCrash(fmt.Sprintf("This post request produced both no results and no resulting apiResponses. []ApiResponse: %#v", *resultPages))
	}
	return resp, nil
}

// GeneratePOSTResponse creates a response that is directly returned to a custom request by the remote.
func GeneratePOSTResponse(respType string, req api.ApiResponse) ([]byte, error) {
	var resp api.ApiResponse
	// Look at filters to figure out what is being requested
	filters := processFilters(&req)
	switch respType {
	case "node":
		r := GeneratePrefilledApiResponse()
		resp = *r
		// resp.Endpoint = "node"
		resp.Entity = "node"
	case "boards", "threads", "posts", "votes", "keys", "truststates":
		localData, dbError := persistence.Read(respType, filters.Fingerprints, filters.Embeds, filters.TimeStart, filters.TimeEnd)
		if dbError != nil {
			return []byte{}, errors.New(fmt.Sprintf("The query coming from the remote caused an error in the local database while trying to respond to this request. Error: %#v\n, Request: %#v\n", dbError, req))
		}
		pages := splitEntitiesToPages(&localData)
		pagesAsApiResponses := convertResponsesToApiResponses(pages)
		finalResponse, err := bakeFinalApiResponse(pagesAsApiResponses)
		// fmt.Printf("%#v", finalResponse)
		if err != nil {
			return []byte{}, errors.New(fmt.Sprintf("An error was encountered while trying to finalise the API response. Error: %#v\n, Request: %#v\n", err, req))
		}
		resp = *finalResponse
		// resp.Endpoint = "entity"
	case "addresses": // Addresses can't do address search by loc/subloc/port. Only time search is available, since addresses don't have fingerprints defined.
		addresses, dbError := persistence.ReadAddresses("", "", 0, filters.TimeStart, filters.TimeEnd, 0, 0, 0)
		var localData api.Response
		localData.Addresses = addresses
		if dbError != nil {
			return []byte{}, errors.New(fmt.Sprintf("The query coming from the remote caused an error in the local database while trying to respond to this request. Error: %#v\n, Request: %#v\n", dbError, req))
		}
		pages := splitEntitiesToPages(&localData)
		pagesAsApiResponses := convertResponsesToApiResponses(pages)
		finalResponse, err := bakeFinalApiResponse(pagesAsApiResponses)
		if err != nil {
			return []byte{}, errors.New(fmt.Sprintf("An error was encountered while trying to finalise the API response. Error: %#v\n, Request: %#v\n", err, req))
		}
		resp = *finalResponse
		resp.Endpoint = "entity"
	}
	// Build the response itself
	resp.Entity = respType
	resp.Timestamp = api.Timestamp(time.Now().Unix())
	// Construct the query, and run an index to determine how many entries we have for the filter.
	jsonResp, err := ConvertApiResponseToJson(&resp)
	if err != nil {
		return []byte{}, errors.New(fmt.Sprintf("The response that was prepared to respond to this query failed to convert to JSON. Error: %#v\n, Request Body: %#v\n", err, req))
	}
	return jsonResp, nil
}

func createBoardIndex(entity *api.Board, pageNum int) api.BoardIndex {
	var entityIndex api.BoardIndex
	entityIndex.Creation = entity.Creation
	entityIndex.Fingerprint = entity.GetFingerprint()
	entityIndex.LastUpdate = entity.LastUpdate
	entityIndex.PageNumber = pageNum
	return entityIndex
}

func createThreadIndex(entity *api.Thread, pageNum int) api.ThreadIndex {
	var entityIndex api.ThreadIndex
	entityIndex.Board = entity.Board
	entityIndex.Creation = entity.Creation
	entityIndex.Fingerprint = entity.GetFingerprint()
	entityIndex.PageNumber = pageNum
	return entityIndex
}

func createPostIndex(entity *api.Post, pageNum int) api.PostIndex {
	var entityIndex api.PostIndex
	entityIndex.Board = entity.Board
	entityIndex.Thread = entity.Thread
	entityIndex.Creation = entity.Creation
	entityIndex.Fingerprint = entity.GetFingerprint()
	entityIndex.PageNumber = pageNum
	return entityIndex
}

func createVoteIndex(entity *api.Vote, pageNum int) api.VoteIndex {
	var entityIndex api.VoteIndex
	entityIndex.Board = entity.Board
	entityIndex.Thread = entity.Thread
	entityIndex.Target = entity.Target
	entityIndex.Creation = entity.Creation
	entityIndex.Fingerprint = entity.GetFingerprint()
	entityIndex.LastUpdate = entity.LastUpdate
	entityIndex.PageNumber = pageNum
	return entityIndex
}

func createKeyIndex(entity *api.Key, pageNum int) api.KeyIndex {
	var entityIndex api.KeyIndex
	entityIndex.Creation = entity.Creation
	entityIndex.Fingerprint = entity.GetFingerprint()
	entityIndex.LastUpdate = entity.LastUpdate
	entityIndex.PageNumber = pageNum
	return entityIndex
}

func createTruststateIndex(entity *api.Truststate, pageNum int) api.TruststateIndex {
	var entityIndex api.TruststateIndex
	entityIndex.Target = entity.Target
	entityIndex.Creation = entity.Creation
	entityIndex.Fingerprint = entity.GetFingerprint()
	entityIndex.LastUpdate = entity.LastUpdate
	entityIndex.PageNumber = pageNum
	return entityIndex
}

// createIndexes creates the index variant of every entity in an api.Response, and puts it back inside one single container for all indexes.
func createIndexes(fullData *[]api.Response) *api.Response {
	fd := *fullData
	var resp api.Response
	if len(fd) > 0 {
		for i, _ := range fd {
			// For each Api.Response page
			if len(fd[i].Boards) > 0 {
				for j, _ := range fd[i].Boards {
					entityIndex := createBoardIndex(&fd[i].Boards[j], i)
					resp.BoardIndexes = append(resp.BoardIndexes, entityIndex)
				}
			}
			if len(fd[i].Threads) > 0 {
				for j, _ := range fd[i].Threads {
					entityIndex := createThreadIndex(&fd[i].Threads[j], i)
					resp.ThreadIndexes = append(resp.ThreadIndexes, entityIndex)
				}
			}
			if len(fd[i].Posts) > 0 {
				for j, _ := range fd[i].Posts {
					entityIndex := createPostIndex(&fd[i].Posts[j], i)
					resp.PostIndexes = append(resp.PostIndexes, entityIndex)
				}
			}
			if len(fd[i].Votes) > 0 {
				for j, _ := range fd[i].Votes {
					entityIndex := createVoteIndex(&fd[i].Votes[j], i)
					resp.VoteIndexes = append(resp.VoteIndexes, entityIndex)
				}
			}
			// Addresses: Address doesn't have an index form. It is its own index.
			if len(fd[i].Keys) > 0 {
				for j, _ := range fd[i].Keys {
					entityIndex := createKeyIndex(&fd[i].Keys[j], i)
					resp.KeyIndexes = append(resp.KeyIndexes, entityIndex)
				}
			}
			if len(fd[i].Truststates) > 0 {
				for j, _ := range fd[i].Truststates {
					entityIndex := createTruststateIndex(&fd[i].Truststates[j], i)
					resp.TruststateIndexes = append(resp.TruststateIndexes, entityIndex)
				}
			}
		}
	}
	return &resp
}

func generateCacheName() (string, error) {
	hash, err := generateRandomHash()
	if err != nil {
		return "", err
	}
	n := fmt.Sprint("cache_", hash)
	return n, nil
}

// CacheResponse is the internal procesing structure for generating caches to be saved to the disk.
type CacheResponse struct {
	cacheName   string
	start       api.Timestamp
	end         api.Timestamp
	entityPages *[]api.Response
	indexPages  *[]api.Response
}

// GenerateCacheResponse responds to a cache generation request. This returns an Api.Response entity with entities, entity indexes, and the cache link that needs to be inserted into the index of the endpoint.
// This has no filters.
func GenerateCacheResponse(respType string, start api.Timestamp, end api.Timestamp) (CacheResponse, error) {
	var resp CacheResponse
	switch respType {
	case "boards", "threads", "posts", "votes", "keys", "truststates":
		localData, dbError := persistence.Read(respType, []api.Fingerprint{}, []string{}, start, end)
		if dbError != nil {
			return resp, errors.New(fmt.Sprintf("This cache generation request caused an error in the local database while trying to respond to this request. Error: %#v\n", dbError))
		}
		entityPages := splitEntitiesToPages(&localData)
		indexes := createIndexes(entityPages)
		indexPages := splitEntityIndexesToPages(indexes)
		cn, err := generateCacheName()
		if err != nil {
			return resp, errors.New(fmt.Sprintf("There was an error in the cache generation request serving. Error: %#v\n", err))
		}
		resp.cacheName = cn
		resp.start = start
		resp.end = end
		resp.indexPages = indexPages
		resp.entityPages = entityPages

	case "addresses":
		addresses, dbError := persistence.ReadAddresses("", "", 0, start, end, 0, 0, 0)
		var localData api.Response
		localData.Addresses = addresses
		if dbError != nil {
			return resp, errors.New(fmt.Sprintf("This cache generation request caused an error in the local database while trying to respond to this request. Error: %#v\n", dbError))
		}
		entityPages := splitEntitiesToPages(&localData)
		cn, err := generateCacheName()
		if err != nil {
			return resp, errors.New(fmt.Sprintf("There was an error in the cache generation request serving. Error: %#v\n", err))
		}
		resp.cacheName = cn
		resp.start = start
		resp.end = end
		resp.entityPages = entityPages

	default:
		return resp, errors.New(fmt.Sprintf("The requested entity type is unknown to the cache generator. Entity type: %s", respType))
	}
	return resp, nil
}

func updateCacheIndex(cacheIndex *api.ApiResponse, cacheData *CacheResponse) {
	// Save the cache link into the index.
	var c api.ResultCache
	c.ResponseUrl = cacheData.cacheName
	c.StartsFrom = cacheData.start
	c.EndsAt = cacheData.end
	cacheIndex.Results = append(cacheIndex.Results, c)
	cacheIndex.Timestamp = api.Timestamp(int64(time.Now().Unix()))
	cacheIndex.Caching.ServedFromCache = true
	cacheIndex.Caching.CacheScope = "day"
	// TODO: How many places am I setting this ".Caching" data?

}

// saveCacheToDisk saves an entire cache's data (entities and indexes, inside a folder named based on the cache name) into the proper location on the disk.
func saveCacheToDisk(entityCacheDir string, cacheData *CacheResponse, respType string) error {
	// Create the index directory.
	cacheDir := fmt.Sprint(entityCacheDir, "/", cacheData.cacheName)
	createPath(cacheDir)
	var indexPages []api.ApiResponse
	var indexDir string
	if respType != "addresses" {
		indexDir = fmt.Sprint(entityCacheDir, "/", cacheData.cacheName, "/index")
		createPath(indexDir)
		indexPages = *convertResponsesToApiResponses(cacheData.indexPages)
	}
	// Convert api.Responses to api.ApiResponses for saving.
	entityPages := *convertResponsesToApiResponses(cacheData.entityPages)
	// Iterate over the data, convert api.ApiResponses to JSON, and save.
	for i, _ := range indexPages {
		indexPages[i].Endpoint = "entity_index"
		indexPages[i].Entity = respType
		indexPages[i].Timestamp = api.Timestamp(int64(time.Now().Unix()))
		indexPages[i].Caching.ServedFromCache = true
		indexPages[i].Caching.CurrentCacheUrl = cacheData.cacheName
		// indexPages[i].Caching.PrevCacheUrl // TODO Pulling this is expensive as heck here. Reconsider the need.
		indexPages[i].Caching.CacheScope = "day"
		// For each index, look at the page number and save the result as that.
		json, _ := ConvertApiResponseToJson(&indexPages[i])
		saveFileToDisk(json, indexDir, fmt.Sprint(indexPages[i].Pagination.CurrentPage, ".json"))
	}
	for i, _ := range entityPages {
		entityPages[i].Endpoint = "entity"
		entityPages[i].Entity = respType
		entityPages[i].Timestamp = api.Timestamp(int64(time.Now().Unix()))
		entityPages[i].Caching.ServedFromCache = true
		entityPages[i].Caching.CurrentCacheUrl = cacheData.cacheName
		// indexPages[i].Caching.PrevCacheUrl // TODO Pulling this is expensive as heck here. Reconsider the need.
		entityPages[i].Caching.CacheScope = "day"
		// For each index, look at the page number and save the result as that.
		json, _ := ConvertApiResponseToJson(&entityPages[i])
		saveFileToDisk(json, cacheDir, fmt.Sprint(entityPages[i].Pagination.CurrentPage, ".json"))
	}
	return nil
}

// CreateCache creates the cache for the given entity type for the given time range.
func CreateCache(respType string, start api.Timestamp, end api.Timestamp) error {
	// - Pull the data from the DB
	// - Look at the cache folder. If there is a cache folder and an index there, save the cache and add to index.
	// - If there is no cache present there, create the index and add it as the first entry.
	cacheData, err := GenerateCacheResponse(respType, start, end)
	if err != nil {
		return errors.New(fmt.Sprintf("Cache creation process encountered an error. Error: %s", err))
	}
	entityCacheDir := fmt.Sprint(globals.CachesLocation, "/", respType)
	// Create the caches dir and the appropriate endpoint if does not exist.
	createPath(entityCacheDir)
	// Save the cache to disk.
	err2 := saveCacheToDisk(entityCacheDir, &cacheData, respType)
	// TODO: above needs to add caching tag, entity and endpoint fields, and the current timestamp.
	if err2 != nil {
		return errors.New(fmt.Sprintf("Cache creation process encountered an error. Error: %s", err2))
	}
	var apiResp api.ApiResponse
	// Look for the index.json in it. If it doesn't exist, create.
	cacheIndexAsJson, err3 := ioutil.ReadFile(fmt.Sprint(entityCacheDir, "/index.json"))
	if err3 != nil && strings.Contains(err3.Error(), "no such file or directory") {
		// The index.json of this cache likely doesn't exist. Create one.
		apiResp = *GeneratePrefilledApiResponse()
	} else if err3 != nil {
		return errors.New(fmt.Sprintf("Cache creation process encountered an error. Error: %s", err3))
	} else {
		// err3 is nil
		json.Unmarshal(cacheIndexAsJson, &apiResp)
	}
	// If the file exists, go through with regular processing.
	updateCacheIndex(&apiResp, &cacheData)
	json, err4 := ConvertApiResponseToJson(&apiResp)
	if err4 != nil {
		return err
	}
	saveFileToDisk(json, entityCacheDir, "index.json")
	return nil
}

// GenerateCaches generates all day caches for all entities and saves them to disk.
func GenerateCaches() {
	now := int64(time.Now().Unix())
	lastCacheGenTs := globals.LastCacheGenerationTimestamp
	lastCacheGenTime := time.Unix(lastCacheGenTs, 0)
	// If more than 24 hours has passed since the last cache generation, generated a new cache for that timeframe.
	if time.Since(lastCacheGenTime) > 24*time.Hour {
		CreateCache("boards", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		CreateCache("threads", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		CreateCache("posts", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		CreateCache("votes", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		CreateCache("addresses", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		CreateCache("keys", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		CreateCache("truststates", api.Timestamp(lastCacheGenTs), api.Timestamp(now))
		// After successfully generating the caches, make the last cache generation timestamp to current.
		globals.LastCacheGenerationTimestamp = now
	}
}
