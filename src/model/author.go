package model

type Author struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Sort string `json:"sort"`
    Link string `json:"link"`
}
