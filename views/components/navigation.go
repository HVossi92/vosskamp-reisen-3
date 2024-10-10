package components

type NavigationLink struct {
	Text string
	Href string
}

var NavigationLinks = []NavigationLink{
	{Text: "Home", Href: "/home"},
	{Text: "Reisen", Href: "/blog"},
	{Text: "Affiliate", Href: "/about-me"},
}
