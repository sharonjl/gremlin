package gremlin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/davecgh/go-spew/spew"
)

func TestOpen(t *testing.T) {
	cli, err := Open("ws://localhost:32768/gremlin")
	if err != nil {
	}
	assert.NotNil(t, cli)
	assert.Nil(t, err)

//	query := `g = TinkerFactory.createModern().traversal();
//g.V().branch(values('name')).option('marko', values('age')).option(none, values('name'))`
	query := `g.V()`
	resp, err := cli.Eval(&EvalInput{Script: query})
	if err != nil {
		t.Error(err)
		return
	}
	spew.Dump(resp)
	query1 := `g.V()`
	resp1, err := cli.Eval(&EvalInput{Script: query1})
	if err != nil {

		t.Error(err)
		return
	}
	var d interface{}
	resp1.Scan(&d)
	spew.Dump(d)
	//cli.Eval(l, query)
	<-time.After(time.Second)
}
