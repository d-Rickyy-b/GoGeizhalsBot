package bot

import "GoGeizhalsBot/internal/geizhals"

var priceagents []PriceAgent

func temp() {
	priceagents = []PriceAgent{{
		ID:   "123",
		Name: "Test",
		Entity: geizhals.Entity{
			Price: 1.99,
			Name:  "Test123",
			URL:   "https://geizhals.at/test123",
			Type:  geizhals.Wishlist,
		},
		NotificationSettings: NotificationSettings{
			NotifyAlways: true,
		},
	},
		{
			ID:   "234",
			Name: "Wishlist2",
			Entity: geizhals.Entity{
				Price: 1.99,
				Name:  "Test234",
				URL:   "https://geizhals.at/test234",
				Type:  geizhals.Wishlist,
			},
			NotificationSettings: NotificationSettings{
				NotifyAlways: true,
			},
		},
		{
			ID:   "456",
			Name: "Wishlist 3",
			Entity: geizhals.Entity{
				Price: 238.00,
				Name:  "Test456",
				URL:   "https://geizhals.at/test456",
				Type:  geizhals.Wishlist,
			},
			NotificationSettings: NotificationSettings{
				NotifyBelow: true,
				BelowPrice:  200.00,
			},
		},
	}
}

func getProductPriceagent() PriceAgent {
	priceagent := PriceAgent{
		ID:   "8923423",
		Name: "Test PriceAgent Product",
		Entity: geizhals.Entity{
			Price: 328.55,
			Name:  "Test Entity Product",
			URL:   "https://example.com/product",
			Type:  geizhals.Product,
		},
		NotificationSettings: NotificationSettings{
			NotifyAlways: true,
		},
	}
	return priceagent
}

func getWishlistPriceagent() PriceAgent {
	priceagent := PriceAgent{
		ID:   "98123112",
		Name: "Test PriceAgent Wishlist",
		Entity: geizhals.Entity{
			Price: 199.99,
			Name:  "Test Entity Wishlist",
			URL:   "https://example.com/wishlist",
			Type:  geizhals.Wishlist,
		},
		NotificationSettings: NotificationSettings{
			NotifyAlways: true,
		},
	}
	return priceagent
}
