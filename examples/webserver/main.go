/*
 * Copyright (c) 2018 Dmytro Lahoza <dmitry@lagoza.name>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/DLag/cachery"
	"os"
)

func router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/orders", handlerOrders)
	r.HandleFunc("/orders/{driver}", handlerOrders)
	r.HandleFunc("/orders/invalidate/{driver}", handlerInvalidateOrders)
	r.HandleFunc("/goods", handlerGoods)
	r.HandleFunc("/goods/{driver}", handlerGoods)
	r.HandleFunc("/goods/invalidate", handlerGoods)
	r.HandleFunc("/goods/invalidate/{driver}", handlerInvalidateGoods)
	return r
}

func handlerInvalidateOrders(w http.ResponseWriter, r *http.Request) {
	if driver, ok := mux.Vars(r)["driver"]; ok {
		cacheName := driver + "_SHORT_CACHE"
		c := cachery.Get(cacheName)
		c.InvalidateAll()
		w.Write([]byte("OK"))
	}
}

func handlerInvalidateGoods(w http.ResponseWriter, r *http.Request) {
	if driver, ok := mux.Vars(r)["driver"]; ok {
		cacheName := driver + "_LONG_CACHE"
		c := cachery.Get(cacheName)
		c.InvalidateAll()
		w.Write([]byte("OK"))
	}
}

func handlerOrders(w http.ResponseWriter, r *http.Request) {
	var o []order
	if driver, ok := mux.Vars(r)["driver"]; ok {
		cacheName := driver + "_SHORT_CACHE"
		c := cachery.Get(cacheName)
		c.Get("", &o, fetcherOrders)
	} else {
		o = getOrders()
	}
	buf, _ := json.Marshal(orders)
	w.Write(buf)
}

func handlerGoods(w http.ResponseWriter, r *http.Request) {
	var g []good
	if driver, ok := mux.Vars(r)["driver"]; ok {
		cacheName := driver + "_LONG_CACHE"
		c := cachery.Get(cacheName)
		c.Get("", &g, fetcherGoods)
	} else {
		g = getGoods()
	}
	buf, _ := json.Marshal(orders)
	w.Write(buf)
}

func main() {
	initRedisCache()
	initInMemCache()
	initInMemNATSCache()
	http.ListenAndServe(os.Args[1], router())
}