package model

// Struct to get coordinates by city name
type WeatherCoordinates struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
	State   string  `json:"state"`
}

// Structs to get weather by coordinates
type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type City struct {
	Id         int32  `json:"id"`
	Name       string `json:"name"`
	Coord      Coord  `json:"coord"`
	Country    string `json:"country"`
	Population int64  `json:"population"`
	Timezone   int32  `json:"timezone"`
	Sunrise    int64  `json:"sunrise"`
	Sunset     int64  `json:"sunset"`
}

type Main struct {
	Temp      float32 `json:"temp"`
	FeelsLike float32 `json:"feels_like"`
	TempMin   float32 `json:"temp_min"`
	TempMax   float32 `json:"temp_max"`
	Pressure  int32   `json:"pressure"`
	SeaLevel  int32   `json:"sea_level"`
	GrndLevel int32   `json:"grnd_level"`
	Humidity  int32   `json:"humidity"`
	TempKf    float32 `json:"temp_kf"`
}

type Weather struct {
	Id          int32  `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Clouds struct {
	All int32 `json:"all"`
}

type Wind struct {
	Speed float32 `json:"speed"`
	Deg   float32 `json:"deg"`
	Gust  float32 `json:"gust"`
}

type Sys struct {
	Pod string `json:"pod"`
}

type List struct {
	Dt         int64     `json:"dt"`
	Main       Main      `json:"main"`
	Weather    []Weather `json:"weather"`
	Clouds     Clouds    `json:"clouds"`
	Wind       Wind      `json:"wind"`
	Visibility int32     `json:"visibility"`
	Pop        float32   `json:"pop"`
	Sys        Sys       `json:"sys"`
	DtTxt      string    `json:"dt_txt"`
}

type WeatherResponse struct {
	City City   `json:"city"`
	List []List `json:"list"`
}
