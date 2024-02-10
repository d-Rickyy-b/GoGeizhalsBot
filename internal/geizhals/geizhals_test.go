package geizhals

import (
	"reflect"
	"testing"
)

func Test_parseGeizhalsURL(t *testing.T) {
	type args struct {
		rawurl     string
		entityType EntityType
	}
	tests := []struct {
		name    string
		args    args
		want    EntityURL
		wantErr bool
	}{
		{
			name: "Wishlist URL DE",
			args: args{
				rawurl:     "https://geizhals.de/?cat=WL-1156092&a=1",
				entityType: Wishlist,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.de/?cat=WL-1156092&a=1",
				CleanURL:     "https://geizhals.de/?cat=WL-1156092",
				Path:         "?cat=WL-1156092",
				Location:     "de",
				EntityID:     -1156092,
				Type:         Wishlist,
			},
			wantErr: false,
		},
		{
			name: "Wishlist URL AT",
			args: args{
				rawurl:     "https://geizhals.at/?cat=WL-1156092",
				entityType: Wishlist,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.at/?cat=WL-1156092",
				CleanURL:     "https://geizhals.at/?cat=WL-1156092",
				Path:         "?cat=WL-1156092",
				Location:     "at",
				EntityID:     -1156092,
				Type:         Wishlist,
			},
			wantErr: false,
		},
		{
			name: "Wishlist URL UK",
			args: args{
				rawurl:     "https://skinflint.co.uk/?cat=WL-1156092",
				entityType: Wishlist,
			},
			want: EntityURL{
				SubmittedURL: "https://skinflint.co.uk/?cat=WL-1156092",
				CleanURL:     "https://skinflint.co.uk/?cat=WL-1156092",
				Path:         "?cat=WL-1156092",
				Location:     "uk",
				EntityID:     -1156092,
				Type:         Wishlist,
			},
			wantErr: false,
		},
		{
			name: "Product URL",
			args: args{
				rawurl:     "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=pl",
				entityType: Product,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=pl",
				CleanURL:     "https://geizhals.de/jabra-elite-85t-a2378831.html",
				Path:         "jabra-elite-85t-a2378831.html",
				Location:     "de",
				EntityID:     2378831,
				Type:         Product,
			},
			wantErr: false,
		},
		{
			name: "Product URL Multiple",
			args: args{
				rawurl:     "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=pl&hloc=de",
				entityType: Product,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=pl&hloc=de",
				CleanURL:     "https://geizhals.de/jabra-elite-85t-a2378831.html",
				Path:         "jabra-elite-85t-a2378831.html",
				Location:     "de",
				EntityID:     2378831,
				Type:         Product,
			},
			wantErr: false,
		},
		{
			name: "Product URL all",
			args: args{
				rawurl:     "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=pl&hloc=de&hloc=at&hloc=uk",
				entityType: Product,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=pl&hloc=de&hloc=at&hloc=uk",
				CleanURL:     "https://geizhals.de/jabra-elite-85t-a2378831.html",
				Path:         "jabra-elite-85t-a2378831.html",
				Location:     "de",
				EntityID:     2378831,
				Type:         Product,
			},
			wantErr: false,
		},
		{
			name: "Product URL wrong hloc",
			args: args{
				rawurl:     "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=fr",
				entityType: Product,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.de/jabra-elite-85t-a2378831.html?hloc=fr",
				CleanURL:     "https://geizhals.de/jabra-elite-85t-a2378831.html",
				Path:         "jabra-elite-85t-a2378831.html",
				Location:     "de",
				EntityID:     2378831,
				Type:         Product,
			},
		},
		{
			name: "Product URL wrong hloc and TLD",
			args: args{
				rawurl:     "https://geizhals.fr/jabra-elite-85t-a2378831.html?hloc=fr",
				entityType: Product,
			},
			want:    EntityURL{},
			wantErr: true,
		},
		{
			name: "New wishlist URL format",
			args: args{
				rawurl:     "https://geizhals.de/wishlists/2564724",
				entityType: Wishlist,
			},
			want: EntityURL{
				SubmittedURL: "https://geizhals.de/wishlists/2564724",
				CleanURL:     "https://geizhals.de/wishlists/2564724",
				Path:         "wishlists/2564724",
				Location:     "de",
				EntityID:     -2564724,
				Type:         Wishlist,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGeizhalsURL(tt.args.rawurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGeizhalsURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseGeizhalsURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
