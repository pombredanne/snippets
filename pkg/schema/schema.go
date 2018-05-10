package schema

type Post struct {
	UserName     string `gorm:"primary_key"`
	Year         int    `gorm:"primary_key"`
	Week         int    `gorm:"primary_key"`
	BodyThisWeek string
	BodyNextWeek string
}

type Subscription struct {
	Subscriber string `gorm:"primary_key"`
	Subscribee string `gorm:"primary_key"`
}

type User struct {
	UserName     string `gorm:"primary_key"`
	RealName     string
	EmailAddress string
}
