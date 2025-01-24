package domain

import "time"

// Price 소수점을 가지는 가격이 많다. 따라서 문자열로 표현한다.
type Price string

// Trade 특정 일자의 트레이드 정보
type Trade struct {
	date         time.Time // e.g. 2025-01-21
	lastPrice    Price
	openingPrice Price
	maxPrice     Price
	minPrice     Price
}
