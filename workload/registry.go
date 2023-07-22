package workload

import (
	"net/url"
	"encoding/json"
	"log"
	"reflect"
	"fmt"
)

type WorkloadRegistry struct {
	argGenFuncMap map[string]func(bool)url.Values
}

func NewWorkloadRegistry() *WorkloadRegistry {
	m := make(map[string]func(bool)url.Values)
	// Hotel Reservation API
	m["hotel_UserHandler"] = UserHandler
	m["hotel_SearchHandler"] = SearchHandler
	m["hotel_RecommendHandler"] = RecommendHandler
	m["hotel_ReservationHandler"] = ReservationHandler

	// Social Network API
	m["sn_ReadHomeTimeline"] = sn_ReadHomeTimeline
	m["sn_ReadUserTimeline"] = sn_ReadUserTimeline
	m["sn_ComposePost"] = sn_ComposePost

	// Leaf API
	m["leaf_Leaf"] = leaf_Leaf
	return &WorkloadRegistry{argGenFuncMap: m}
}

func (r *WorkloadRegistry) GetGeneratorFunction(name string) func(bool)url.Values {
	if v, ok := r.argGenFuncMap[name]; ok {
		return v
	}
	log.Println("Returning nil for", name)
	return nil
}

func prepareArg(arg interface{}, is_original bool) string {
	if is_original && reflect.TypeOf(arg).Kind() == reflect.String {
		return fmt.Sprintf("%s", arg)
	}
	bytes, _ := json.Marshal(arg)
	return string(bytes)
}
