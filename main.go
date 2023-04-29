package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/ypapax/logrus_conf"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	if err := func() error {
		if err := logrus_conf.PrepareFromEnv("date_hole"); err != nil {
			return errors.WithStack(err)
		}
		var filePath, timeFormat string
		var minHoleMinutes, maxItemsToShow int
		flag.StringVar(&filePath, "file", "", "path of a target file")
		flag.StringVar(&timeFormat, "time_format", time.UnixDate, "time format")
		flag.IntVar(&minHoleMinutes, "time_hole_minutes", 10, "minimum amount of minutes to find time holes")
		flag.IntVar(&maxItemsToShow, "show_limit", 5, "maximum filtered items to show")
		flag.Parse()
		lf := logrus.WithField("filePath", filePath).WithField("timeFormat", timeFormat).WithField("minHoleMinutes", minHoleMinutes)
		if len(filePath) == 0 {
			return errors.Errorf("missing filePath")
		}
		b, err := os.ReadFile(filePath)
		if err != nil {
			return errors.WithStack(err)
		}
		all, chronological, longestLast := FindFarDates(b, minHoleMinutes, timeFormat, maxItemsToShow)
		if len(all) == 0 {
			lf.Warnf("no dates found")
			return nil
		}
		lf = lf.
			WithField("analyzed-period", fmt.Sprintf("%+v-%+v", all[0].Time, all[len(all)].Time)).
			WithField("analyzed-count", len(all))
		Print(chronological, lf, "chronological", timeFormat)
		Print(longestLast, lf, "longest", timeFormat)
		return nil
	}(); err != nil {
		logrus.Errorf("%+v", err)
	}
}

type Hole struct {
	Time time.Time
	//TimeLine string
	//TextAfter string
	NextTime time.Time
	NextSpace time.Duration
}

func Print(hh []Hole, l *logrus.Entry, label string, format string) {
	lf := l.WithField("label", label)
	for _, h := range hh {
		lf.Infof("%+v : '%+v' - '%+v'", h.NextSpace, h.Time.Format(format), h.NextTime.Format(format))
	}
}

func FindFarDates(b []byte, minHoleMinutes int, timeLayout string, limitShowMax int) (allDates []Hole, filteredChronological []Hole, filteredHeaviestLast []Hole) {
	targetMinDur := time.Duration(minHoleMinutes) * time.Minute
	lines := strings.Split(string(b), "\n")
	var dd []time.Time
	for _, l := range lines {
		l = strings.TrimSpace(l)
		t, err := time.Parse(timeLayout, l)
		if err != nil {
			continue
		}
		dd = append(dd, t)
	}
	for i, d := range dd {
		if i == len(dd) - 1 {
			break
		}
		next := dd[i+1]
		space := next.Sub(d)
		h := Hole{Time: d, NextSpace: space, NextTime: next}
		allDates = append(allDates, h)
		if space > targetMinDur {
			filteredChronological = append(filteredChronological, h)
		}
	}
	filteredHeaviestLast = append(filteredChronological)
	sort.Slice(filteredHeaviestLast, func(i, j int) bool {
		return filteredHeaviestLast[i].NextSpace < filteredHeaviestLast[j].NextSpace
	})
	if len(filteredChronological) > limitShowMax {
		filteredChronological = filteredChronological[len(filteredChronological)-limitShowMax:]
	}
	if len(filteredHeaviestLast) > limitShowMax {
		filteredHeaviestLast = filteredHeaviestLast[len(filteredHeaviestLast)-limitShowMax:]
	}
	return allDates, filteredChronological, filteredHeaviestLast
}
