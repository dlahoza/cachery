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
	"fmt"
	"math/rand"
	"time"
)

type order struct {
	Id   int
	User int
	Good int
}

type good struct {
	Id   int
	Name string
}

var (
	goods  []good
	orders []order
)

func init() {
	for i := 0; i < 10; i++ {
		id := rand.Int()
		goods = append(goods, good{Id: id, Name: fmt.Sprintf("Good %d", id)})
	}
	for i := 0; i < 20; i++ {
		orders = append(orders, order{Id: rand.Int(), User: rand.Int(), Good: rand.Int()})
	}
}

func getGoods() []good {
	time.Sleep(time.Second)
	return goods
}

func getOrders() []order {
	time.Sleep(time.Second)
	return orders
}

func fetcherGoods(key interface{}) (interface{}, error) {
	return getGoods(), nil
}

func fetcherOrders(key interface{}) (interface{}, error) {
	return getOrders(), nil
}
