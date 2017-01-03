// Compile top-tags-per-annum statistics from last.fm history.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shkh/lastfm-go/lastfm"
)

const ()

var (
	apiKey             = flag.String("api-key", "", "Last.fm API key")
	apiSecret          = flag.String("api-secret", "", "Last.fm API secret")
	user               = flag.String("user", "", "Last.fm user name")
	tracksLimit        = flag.Int("tracks-limit", -1, "Maximum number of tracks to process")
	tracksPerIndicator = flag.Int("tracks-per-indicator", 1000, "Show indicator for every K tracks processed")
	topThreshold       = flag.Int("top-threshold", 5, "Top tags to print per year")
	stateFile          = flag.String("state", "state.json", "Name of the state file")
)

func main() {
	flag.Parse()

	s := &State{}
	s.load()
	s.update()
	s.save()
	s.printStats()
}

// Note that the key for AnnualCounts and AnnualPlays is a string, so
// that Golang's JSON encoder can dump it.  What a moronic limitation.
//
// Also the decision to overload the meaning of a name's first letter
// to both export the name and to dump as JSON sucks when I develop a
// program.

type TagAndCount struct {
	Name  string
	Count int
}
type TagCounts map[string]*TagAndCount
type AnnualCounts map[string]TagCounts

type State struct {
	ArtistTag      map[string]string
	Alltime        AnnualCounts
	AnnualPlays    map[string]int
	MinYear        int
	MaxYear        int
	MostRecentPlay time.Time
}

func (s *State) update() {
	if *apiKey == "" {
		log.Fatal("Need API key")
	}
	if *apiSecret == "" {
		log.Fatal("Need API secret")
	}
	if *user == "" {
		log.Fatal("Need last.fm user name")
	}

	api := lastfm.New(*apiKey, *apiSecret)

	token, _ := api.GetToken()
	api.GetAuthTokenUrl(token)
	api.LoginWithToken(token)

	doEachTrack(api, s.MostRecentPlay, func(artistName string, trackName string, playTime time.Time) {
		if playTime.After(s.MostRecentPlay) {
			s.MostRecentPlay = playTime
		}

		if s.ArtistTag[artistName] == "" {
			s.ArtistTag[artistName] = getTopTag(api, artistName)
		}
		tag := s.ArtistTag[artistName]

		year := playTime.Year()
		if year < s.MinYear {
			s.MinYear = year
		}
		if year > s.MaxYear {
			s.MaxYear = year
		}

		yearString := strconv.Itoa(year)

		s.AnnualPlays[yearString]++

		if s.Alltime[yearString] == nil {
			s.Alltime[yearString] = make(TagCounts)
		}
		if s.Alltime[yearString][tag] == nil {
			s.Alltime[yearString][tag] = &TagAndCount{
				Name:  tag,
				Count: 1,
			}
		} else {
			s.Alltime[yearString][tag].Count++
		}
	})
}

func (s *State) save() {
	f, err := os.Create(*stateFile)
	if err != nil {
		log.Printf("Couldn't save state: %v\n", err)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s); err != nil {
		log.Printf("Couldn't encode state: %v\n", err)
	}
}

func (s *State) defaultInit() {
	s.ArtistTag = make(map[string]string)
	s.Alltime = make(AnnualCounts)
	s.AnnualPlays = make(map[string]int)
	s.MinYear = time.Now().Year()
	s.MaxYear = time.Now().Year()
	s.MostRecentPlay = time.Unix(0, 0)
}

func (s *State) load() {
	f, err := os.Open(*stateFile)
	if os.IsNotExist(err) {
		s.defaultInit()
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(s); err != nil {
		log.Fatal(err)
	}
}

func (s *State) printStats() {
	for year := s.MinYear; year <= s.MaxYear; year++ {
		yearString := strconv.Itoa(year)
		counts := s.Alltime[yearString]
		fmt.Printf("%d: ", year)
		top := annualTopTags(counts)
		for i, tc := range top {
			if i > 0 {
				fmt.Printf(", ")
			}
			pct := float64(tc.Count) / float64(s.AnnualPlays[yearString]) * 100
			fmt.Printf("%s (%.1f)", tc.Name, pct)
		}
		fmt.Printf("\n")
	}
}

type trackProcessor func(artistName string, trackName string, playTime time.Time)

func doEachTrack(api *lastfm.Api, from time.Time, fn trackProcessor) {
	pageNumber := 1

	if *tracksLimit == 0 {
		return
	}
	tracksProcessed := 0

	recentTracks := fetchPage(api, from, pageNumber)
	for pageNumber <= recentTracks.TotalPages {
		for _, track := range recentTracks.Tracks {
			// Old last.fm entries apparently don't have a date string.
			if track.Date.Date == "" {
				continue
			}
			playTime, err := time.Parse("02 Jan 2006, 15:04", track.Date.Date)
			if err != nil {
				log.Fatal(err)
			}
			fn(track.Artist.Name, track.Name, playTime)
			tracksProcessed++
			if *tracksPerIndicator > 0 && (tracksProcessed%*tracksPerIndicator) == 0 {
				log.Printf("Considered %7d/%7d tracks\n", tracksProcessed, recentTracks.Total)
			}
			if tracksProcessed == *tracksLimit {
				return
			}
		}
		pageNumber++
		recentTracks = fetchPage(api, from, pageNumber)
	}
}

func fetchPage(api *lastfm.Api, from time.Time, pageNumber int) *lastfm.UserGetRecentTracks {
	recentTracks, err := api.User.GetRecentTracks(lastfm.P{
		"limit": 200,
		"user":  *user,
		"page":  pageNumber,
		"from":  from.Unix(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return &recentTracks
}

func getTopTag(api *lastfm.Api, artistName string) string {
	topTags, err := api.Artist.GetTopTags(lastfm.P{
		"artist": artistName,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(topTags.Tags) == 0 {
		return "(no tag)"
	}
	return canonicalizeTag(topTags.Tags[0].Name)
}

func canonicalizeTag(rawTag string) string {
	return strings.ToLower(rawTag)
}

type byTagCount struct {
	counts []*TagAndCount
}

func (b byTagCount) Len() int           { return len(b.counts) }
func (b byTagCount) Less(i, j int) bool { return b.counts[i].Count > b.counts[j].Count }
func (b byTagCount) Swap(i, j int)      { b.counts[i], b.counts[j] = b.counts[j], b.counts[i] }

func annualTopTags(counts TagCounts) []*TagAndCount {
	ordered := make([]*TagAndCount, len(counts))
	i := 0
	for _, tc := range counts {
		ordered[i] = tc
		i++
	}

	sort.Sort(byTagCount{ordered})

	k := *topThreshold
	if len(ordered) < k {
		k = len(ordered)
	}
	return ordered[0:k]
}
