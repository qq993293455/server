package self_ip

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelf(t *testing.T) {
	r := require.New(t)

	w, l := Self(&IpServiceAddrS{
		Lan: "10.23.20.53:25000",
		Wan: "",
	})
	r.NotEqual(w, "")
	r.NotEqual(l, "")
	fmt.Println(w, l)
}
