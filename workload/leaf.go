package workload

import (
	"math/rand"
	"net/url"
)

func leaf_Leaf(is_original bool) url.Values {
	a := rand.Intn(100)
	data := url.Values{}
	data.Add("a", prepareArg(a, is_original))
	return data
}