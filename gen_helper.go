// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var (
	oldGenFiles = make(map[string]bool)
)

func main() {
	clearOldGenFiles()
	genIncludeFiles()
	printOldGenFiles()
}

func clearOldGenFiles() {
	ss, err := filepath.Glob("z_*.c")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(ss); i++ {
		ioutil.WriteFile(ss[i], []byte("#error file removed!!!\n"), 0666)
		oldGenFiles[ss[i]] = true
	}
}

func genIncludeFiles() {
	ss := parseCMakeListsTxt("internal/libwebp/CMakeLists.txt", "WEBP_SRC", ".c")
	for i := 0; i < len(ss); i++ {
		relpath := ss[i][2:] // drop `./`
		newname := "z_libwebp_" + strings.Replace(relpath, "/", "_", -1)

		ioutil.WriteFile(newname, []byte(fmt.Sprintf(
			`// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "./internal/libwebp/%s"
`, relpath,
		)), 0666)

		delete(oldGenFiles, newname)
	}
}

func printOldGenFiles() {
	if len(oldGenFiles) == 0 {
		return
	}
	fmt.Printf("Removed Files:\n")
	for k, _ := range oldGenFiles {
		fmt.Printf("%s\n", k)
	}
	fmt.Printf("Total %d\n", len(oldGenFiles))
}

func parseCMakeListsTxt(filename, varname, ext string) (ss []string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	br := bufio.NewReader(bytes.NewReader(data))

	// find set($varname
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(string(line), "set("+varname) {
			break
		}
	}

	// read $varname, end with `)`
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(string(line), ")") {
			break
		}
		switch v := strings.TrimSpace(string(line)); {
		case strings.HasPrefix(v, `${`): // parse ${?}
			ss = append(ss, parseCMakeListsTxt(filename, v[2:len(v)-3], ext)...)
		case strings.HasSuffix(v, ext): // *.ext
			ss = append(ss, v)
		}
	}
	return
}
