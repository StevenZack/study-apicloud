package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/StevenZack/tools"
	"gopkg.in/mgo.v2"
)

type Video struct {
	Title, Url string
}

func main() {
	http.HandleFunc("/upload", upload)
	http.Handle("/pub/", http.StripPrefix("/pub/", http.FileServer(http.Dir("pub"))))
	http.HandleFunc("/vlist", vlist)
	e := http.ListenAndServe(":9898", nil)
	if e != nil {
		fmt.Println(e)
		return
	}
}
func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	title := handleStr(r, "title")
	rpath := "pub/" + tools.NewToken()
	fis := r.MultipartForm.File["video"]
	for _, v := range fis {
		fi, e := v.Open()
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		os.MkdirAll("pub/", 0755)
		_, e = os.Stat(rpath)
		if e == nil {
			rpath = "pub/" + tools.NewNumToken()
		}
		fo, e := os.OpenFile(rpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(fo, fi)
		fo.Close()
		break
	}
	s, e := mgo.Dial("127.0.0.1")
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	defer s.Close()
	e = s.DB("study").C("videos").Insert(Video{Title: title, Url: "http://101.200.54.63:9898/" + rpath})
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "OK")
}
func vlist(w http.ResponseWriter, r *http.Request) {
	var backData = struct {
		Status string
		Videos []Video
	}{}
	s, e := mgo.Dial("127.0.0.1")
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	defer s.Close()
	e = s.DB("study").C("videos").Find(nil).All(&backData.Videos)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	b, e := json.Marshal(backData)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
func handleV(r *http.Request, key string) string {
	vs := r.MultipartForm.File[key]
	if len(vs) > 0 {
		fi, e := vs[0].Open()
		if e != nil {
			fmt.Println(e)
			return ""
		}
		str := tools.ReadAll(fi)
		fmt.Println(str)
		fi.Close()
		return str
	}
	return ""
}
func handleStr(r *http.Request, key string) string {
	openId := handleV(r, key)
	if openId == "" {
		if r.MultipartForm.Value != nil {
			fmt.Println("multi != nil")
			if v, ok := r.MultipartForm.Value[key]; ok {
				if len(v) > 0 {
					fmt.Println("len(v) > 0")
					openId = v[0]
				}
			}
		} else {
			fmt.Println("r.FormValue")
			openId = r.FormValue("openId")
		}
	}
	return openId
}
