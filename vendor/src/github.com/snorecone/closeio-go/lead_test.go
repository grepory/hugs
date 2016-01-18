package closeio

import (
	"log"
	"os"
	"testing"
)

func TestLeadList(t *testing.T) {

	key := os.Getenv("CLOSEIO_KEY")
	closeAPI := New(key)
	leads, err := closeAPI.Leads(nil)
	if err != nil {
		log.Println(err)
	}
	log.Println(leads.Data[0])

}
func TestLeadQuery(t *testing.T) {
	key := os.Getenv("CLOSEIO_KEY")
	closeAPI := New(key)
	ls := &LeadSearch{
		Query: "name:Raleigh",
	}

	leads, err := closeAPI.Leads(ls)
	if err != nil {
		log.Println(err)
	}
	log.Println(leads.Data[0])

}

//func TestLeadCreate(t *testing.T) {
//	key := os.Getenv("CLOSEIO_KEY")
//	closeAPI := New(key)
//	lead := Lead{
//			Name: "Bluth Company",
//			Url: "http://thebluthcompany.tumblr.com/",
//			Description: "Best. Show. Ever.",
//			//StatusId: "stat_1ZdiZqcSIkoGVnNOyxiEY58eTGQmFNG3LPlEVQ4V7Nk",
//			Contacts: []Contact{
//				Contact{
//					Name: "Gob",
//					Title: "Sr. Vice President",
//					Emails: []Email{
//						Email{
//							Type: "office",
//							Email: "gob@example.com",
//						},
//					},
//					Phones: []Phone{
//						Phone{
//							Type: "office",
//							Phone: "8004445555",
//						},
//					},
//				},
//			},
//			Custom: map[string]string{
//				"Source": "Website contact form",
//				"Transportation": "Segway",
//			},
//			Addresses: []Address{
//				Address{
//					Label: "business",
//					Address1: "747 Howard St",
//					Address2: "Room 3",
//					City: "San Francisco",
//					State: "CA",
//					Zipcode: "94103",
//					Country:"US",
//				},
//			},
//		}
//		leadresp, err := closeAPI.Create(&lead)
//		if err != nil {
//			log.Println(err)
//		}
//		log.Println(leadresp)
//
//}
