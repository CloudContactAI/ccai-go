package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudcontactai/ccai-go/src/pkg/brands"
	"github.com/cloudcontactai/ccai-go/src/pkg/campaigns"
	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
	"github.com/cloudcontactai/ccai-go/src/pkg/webhook"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

var passed, failed int

func run(name string, fn func() error) {
	err := fn()
	if err != nil {
		fmt.Printf("  FAIL [%s]: %v\n", name, err)
		failed++
	} else {
		fmt.Printf("  PASS [%s]\n", name)
		passed++
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "ERROR: required env var %s is not set\n", key)
		os.Exit(2)
	}
	return v
}

// hmacSHA256Base64 computes Base64(HMAC-SHA256(secret, message))
func hmacSHA256Base64(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

// writeTempPNG writes a 1×1 PNG to a temp file and returns the path.
func writeTempPNG() (string, error) {
	const pngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwADhQGAWjR9awAAAABJRU5ErkJggg=="
	decoded, err := base64.StdEncoding.DecodeString(pngB64)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "ccai_test_*.png")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(decoded); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	// ── Validate required env vars ──────────────────────────────────────────
	clientID := mustEnv("CCAI_CLIENT_ID")
	apiKey := mustEnv("CCAI_API_KEY")
	phone1 := mustEnv("CCAI_TEST_PHONE")
	phone2 := mustEnv("CCAI_TEST_PHONE_2")
	phone3 := mustEnv("CCAI_TEST_PHONE_3")
	email1 := mustEnv("CCAI_TEST_EMAIL")
	email2 := mustEnv("CCAI_TEST_EMAIL_2")
	email3 := mustEnv("CCAI_TEST_EMAIL_3")
	firstName1 := mustEnv("CCAI_TEST_FIRST_NAME")
	lastName1 := mustEnv("CCAI_TEST_LAST_NAME")
	firstName2 := mustEnv("CCAI_TEST_FIRST_NAME_2")
	lastName2 := mustEnv("CCAI_TEST_LAST_NAME_2")
	firstName3 := mustEnv("CCAI_TEST_FIRST_NAME_3")
	lastName3 := mustEnv("CCAI_TEST_LAST_NAME_3")
	webhookURL := mustEnv("WEBHOOK_URL")

	// ── Create client ────────────────────────────────────────────────────────
	client, err := ccai.NewClient(ccai.Config{
		ClientID:           clientID,
		APIKey:             apiKey,
		UseTestEnvironment: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create client: %v\n", err)
		os.Exit(2)
	}

	fmt.Println("==============================================")
	fmt.Println("  CCAI Go SDK Integration Tests")
	fmt.Println("==============================================")

	// ── Pre-create temp PNG for MMS tests ────────────────────────────────────
	pngPath, err := writeTempPNG()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create temp PNG: %v\n", err)
		os.Exit(2)
	}
	defer os.Remove(pngPath)
	// Derive absolute path for Docker usage
	pngPath, _ = filepath.Abs(pngPath)

	// ─────────────────────────────────────────────────────────────────────────
	// SMS TESTS (1–6)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- SMS ---")

	// 1. SMS.SendSingle
	run("01 SMS.SendSingle", func() error {
		_, err := client.SMS.SendSingle(firstName1, lastName1, phone1, "Hello from Go SDK!", "Go Test", "", "", nil)
		return err
	})

	// 2. SMS.Send — 1 recipient
	run("02 SMS.Send (1 recipient)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		_, err := client.SMS.Send(accounts, "Hello 1 recipient!", "Go Test", "", nil)
		return err
	})

	// 3. SMS.Send — 2 recipients
	run("03 SMS.Send (2 recipients)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
		}
		_, err := client.SMS.Send(accounts, "Hello 2 recipients!", "Go Test", "", nil)
		return err
	})

	// 4. SMS.Send — 3 recipients
	run("04 SMS.Send (3 recipients)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
			{FirstName: firstName3, LastName: lastName3, Phone: phone3},
		}
		_, err := client.SMS.Send(accounts, "Hello 3 recipients!", "Go Test", "", nil)
		return err
	})

	// 5. SMS.Send with Data field
	run("05 SMS.Send with data", func() error {
		accounts := []sms.Account{
			{
				FirstName: firstName1,
				LastName:  lastName1,
				Phone:     phone1,
				Data:      map[string]string{"city": "Miami", "offer": "20% off"},
			},
		}
		_, err := client.SMS.Send(accounts, "Hello from ${city}! Claim your ${offer}.", "Go Test Data", "", nil)
		return err
	})

	// 6. SMS.Send with MessageData field
	run("06 SMS.Send with messageData", func() error {
		accounts := []sms.Account{
			{
				FirstName:   firstName1,
				LastName:    lastName1,
				Phone:       phone1,
				MessageData: `{"trackingId":"abc123"}`,
			},
		}
		_, err := client.SMS.Send(accounts, "Hello with messageData!", "Go Test MsgData", "", nil)
		return err
	})

	// ─────────────────────────────────────────────────────────────────────────
	// MMS TESTS (7–17)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- MMS ---")

	var signedURLResp *sms.SignedURLResponse
	mmsDepFailed := false

	// 7. MMS.GetSignedUploadURL
	run("07 MMS.GetSignedUploadURL", func() error {
		resp, err := client.MMS.GetSignedUploadURL("test_image.png", "image/png", "", true)
		if err != nil {
			mmsDepFailed = true
			return err
		}
		signedURLResp = resp
		if resp.SignedS3URL == "" {
			mmsDepFailed = true
			return fmt.Errorf("signedS3Url is empty")
		}
		return nil
	})

	// 8. MMS.UploadImageToSignedURL
	run("08 MMS.UploadImageToSignedURL", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		ok, err := client.MMS.UploadImageToSignedURL(signedURLResp.SignedS3URL, pngPath, "image/png")
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("upload returned false")
		}
		return nil
	})

	// 9. MMS.SendSingle
	run("09 MMS.SendSingle", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		_, err := client.MMS.SendSingle(signedURLResp.FileKey, firstName1, lastName1, phone1, "MMS single!", "Go MMS Test", "", "", nil, false)
		return err
	})

	// 10. MMS.Send — 1 recipient
	run("10 MMS.Send (1 recipient)", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		_, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS 1 recipient!", "Go MMS Test", "", nil, false)
		return err
	})

	// 11. MMS.Send — 2 recipients
	run("11 MMS.Send (2 recipients)", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
		}
		_, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS 2 recipients!", "Go MMS Test", "", nil, false)
		return err
	})

	// 12. MMS.Send — 3 recipients
	run("12 MMS.Send (3 recipients)", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
			{FirstName: firstName3, LastName: lastName3, Phone: phone3},
		}
		_, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS 3 recipients!", "Go MMS Test", "", nil, false)
		return err
	})

	// 13. MMS.Send with Data field
	run("13 MMS.Send with data", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{
				FirstName: firstName1,
				LastName:  lastName1,
				Phone:     phone1,
				Data:      map[string]string{"product": "Widget"},
			},
		}
		_, err := client.MMS.Send(signedURLResp.FileKey, accounts, "Check out ${product}!", "Go MMS Data", "", nil, false)
		return err
	})

	// 14. MMS.Send with MessageData field
	run("14 MMS.Send with messageData", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{
				FirstName:   firstName1,
				LastName:    lastName1,
				Phone:       phone1,
				MessageData: `{"campaignId":"mms-test-001"}`,
			},
		}
		_, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS with messageData!", "Go MMS MsgData", "", nil, false)
		return err
	})

	// 15. MMS.CheckFileUploaded
	run("15 MMS.CheckFileUploaded", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		_, err := client.MMS.CheckFileUploaded(signedURLResp.FileKey)
		return err
	})

	// 16. MMS.SendWithImage (fresh upload)
	run("16 MMS.SendWithImage (fresh upload)", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		_, err := client.MMS.SendWithImage(pngPath, "image/png", accounts, "MMS with image!", "Go MMS Image", "", nil, true)
		return err
	})

	// 17. MMS.SendWithImage (cached — same image, should reuse S3 key)
	run("17 MMS.SendWithImage (cached)", func() error {
		if mmsDepFailed {
			return fmt.Errorf("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		_, err := client.MMS.SendWithImage(pngPath, "image/png", accounts, "MMS cached image!", "Go MMS Cache", "", nil, true)
		return err
	})

	// ─────────────────────────────────────────────────────────────────────────
	// EMAIL TESTS (18–22)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Email ---")

	senderEmail := "noreply@cloudcontactai.com"
	senderName := "CCAI Test"
	replyEmail := "noreply@cloudcontactai.com"

	// 18. Email.SendSingle
	run("18 Email.SendSingle", func() error {
		_, err := client.Email.SendSingle(
			firstName1, lastName1, email1,
			"Go SDK Test Email", "<p>Hello from Go SDK!</p>", "",
			senderEmail, replyEmail, senderName, "Go Email Test",
			nil,
		)
		return err
	})

	// 19. Email.Send — 1 recipient
	run("19 Email.Send (1 recipient)", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
		}
		_, err := client.Email.Send(accounts, "Go SDK Email 1", "<p>Hello 1!</p>", senderEmail, replyEmail, senderName, "Go Email Test", nil)
		return err
	})

	// 20. Email.Send — 2 recipients
	run("20 Email.Send (2 recipients)", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
			{FirstName: firstName2, LastName: lastName2, Email: email2},
		}
		_, err := client.Email.Send(accounts, "Go SDK Email 2", "<p>Hello 2!</p>", senderEmail, replyEmail, senderName, "Go Email Test", nil)
		return err
	})

	// 21. Email.Send — 3 recipients
	run("21 Email.Send (3 recipients)", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
			{FirstName: firstName2, LastName: lastName2, Email: email2},
			{FirstName: firstName3, LastName: lastName3, Email: email3},
		}
		_, err := client.Email.Send(accounts, "Go SDK Email 3", "<p>Hello 3!</p>", senderEmail, replyEmail, senderName, "Go Email Test", nil)
		return err
	})

	// 22. Email.SendCampaign (direct struct)
	run("22 Email.SendCampaign", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
			{FirstName: firstName2, LastName: lastName2, Email: email2},
		}
		campaign := &ccai.EmailCampaign{
			Subject:      "Go SDK Campaign Test",
			Title:        "Go Email Campaign",
			Message:      "<p>Campaign email from Go SDK!</p>",
			SenderEmail:  senderEmail,
			ReplyEmail:   replyEmail,
			SenderName:   senderName,
			Accounts:     accounts,
			CampaignType: "EMAIL",
			AddToList:    "noList",
			ContactInput: "accounts",
			FromType:     "single",
			Senders:      []interface{}{},
		}
		_, err := client.Email.SendCampaign(campaign, nil)
		return err
	})

	// ─────────────────────────────────────────────────────────────────────────
	// WEBHOOK TESTS (23–29)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Webhook ---")

	secret := "test-webhook-secret-go"
	var registeredWebhookID string

	// 23. Webhook.Register
	run("23 Webhook.Register", func() error {
		resp, err := client.Webhook.Register(webhook.WebhookConfig{
			URL:    webhookURL,
			Secret: &secret,
		})
		if err != nil {
			return err
		}
		switch v := resp.ID.(type) {
		case float64:
			registeredWebhookID = fmt.Sprintf("%.0f", v)
		case string:
			registeredWebhookID = v
		default:
			registeredWebhookID = fmt.Sprintf("%v", v)
		}
		if registeredWebhookID == "" || registeredWebhookID == "0" {
			return fmt.Errorf("webhook ID is empty after register")
		}
		return nil
	})

	// 24. Webhook.List
	run("24 Webhook.List", func() error {
		hooks, err := client.Webhook.List()
		if err != nil {
			return err
		}
		if len(hooks) == 0 {
			return fmt.Errorf("expected at least one webhook, got 0")
		}
		return nil
	})

	// 25. Webhook.Update
	run("25 Webhook.Update", func() error {
		if registeredWebhookID == "" {
			return fmt.Errorf("no webhook ID from test 23")
		}
		updatedSecret := "updated-secret-go"
		_, err := client.Webhook.Update(registeredWebhookID, webhook.WebhookConfig{
			URL:    webhookURL + "?updated=1",
			Secret: &updatedSecret,
		})
		return err
	})

	// 26. Webhook.VerifySignature — valid
	run("26 Webhook.VerifySignature (valid)", func() error {
		eventHash := "abc123eventHash"
		sig := hmacSHA256Base64(secret, clientID+":"+eventHash)
		ok := client.Webhook.VerifySignature(sig, clientID, eventHash, secret)
		if !ok {
			return fmt.Errorf("expected valid signature to return true")
		}
		return nil
	})

	// 27. Webhook.VerifySignature — invalid
	run("27 Webhook.VerifySignature (invalid)", func() error {
		ok := client.Webhook.VerifySignature("invalidsig==", clientID, "somehash", secret)
		if ok {
			return fmt.Errorf("expected invalid signature to return false")
		}
		return nil
	})

	// 28. Webhook.ParseEvent
	run("28 Webhook.ParseEvent", func() error {
		payload := []byte(`{"eventType":"message.sent","data":{"to":"+15005550001"},"eventHash":"abc123"}`)
		event, err := webhook.ParseEvent(payload)
		if err != nil {
			return err
		}
		if event.EventType == "" {
			return fmt.Errorf("eventType is empty after ParseEvent")
		}
		return nil
	})

	// 29. Webhook.Delete
	run("29 Webhook.Delete", func() error {
		if registeredWebhookID == "" {
			return fmt.Errorf("no webhook ID from test 23")
		}
		_, err := client.Webhook.Delete(registeredWebhookID)
		return err
	})

	// ─────────────────────────────────────────────────────────────────────────
	// CONTACT TESTS (30–31)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Contact ---")

	// 30. Contact.SetDoNotText(true)
	run("30 Contact.SetDoNotText(true)", func() error {
		_, err := client.Contact.SetDoNotText(true, "", phone1)
		return err
	})

	// 31. Contact.SetDoNotText(false)
	run("31 Contact.SetDoNotText(false)", func() error {
		_, err := client.Contact.SetDoNotText(false, "", phone1)
		return err
	})

	// ─────────────────────────────────────────────────────────────────────────
	// BRANDS TESTS (32–36)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Brands ---")

	var createdBrandID int64

	// 32. Brands.Create
	run("32 Brands.create", func() error {
		resp, err := client.Brands.Create(brands.BrandRequest{
			LegalCompanyName: strPtr("Test Company LLC"),
			EntityType:       strPtr("PRIVATE_PROFIT"),
			TaxId:            strPtr("123456789"),
			TaxIdCountry:     strPtr("US"),
			Country:          strPtr("US"),
			VerticalType:     strPtr("TECHNOLOGY"),
			WebsiteUrl:       strPtr("https://example.com"),
			Street:           strPtr("123 Main St"),
			City:             strPtr("Miami"),
			State:            strPtr("FL"),
			PostalCode:       strPtr("33101"),
			ContactFirstName: strPtr(firstName1),
			ContactLastName:  strPtr(lastName1),
			ContactEmail:     strPtr(email1),
			ContactPhone:     strPtr(phone1),
		})
		if err != nil {
			return err
		}
		createdBrandID = resp.ID
		if createdBrandID == 0 {
			return fmt.Errorf("brandId is 0 after create")
		}
		return nil
	})

	// 33. Brands.Get
	run("33 Brands.get", func() error {
		if createdBrandID == 0 {
			return fmt.Errorf("no brandId from test 32")
		}
		_, err := client.Brands.Get(createdBrandID)
		return err
	})

	// 34. Brands.List
	run("34 Brands.list", func() error {
		_, err := client.Brands.List()
		return err
	})

	// 35. Brands.Update
	run("35 Brands.update", func() error {
		if createdBrandID == 0 {
			return fmt.Errorf("no brandId from test 32")
		}
		_, err := client.Brands.Update(createdBrandID, brands.BrandRequest{
			City: strPtr("Orlando"),
		})
		return err
	})

	// 36. Brands.Delete
	run("36 Brands.delete", func() error {
		if createdBrandID == 0 {
			return fmt.Errorf("no brandId from test 32")
		}
		return client.Brands.Delete(createdBrandID)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// CAMPAIGNS TESTS (37–42)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Campaigns ---")

	var campaignBrandID int64
	var createdCampaignID int64

	// 37. Campaign setup — create a brand to associate with campaign
	run("37 Campaign setup — Brands.create", func() error {
		resp, err := client.Brands.Create(brands.BrandRequest{
			LegalCompanyName: strPtr("Campaign Test LLC"),
			EntityType:       strPtr("PRIVATE_PROFIT"),
			TaxId:            strPtr("987654321"),
			TaxIdCountry:     strPtr("US"),
			Country:          strPtr("US"),
			VerticalType:     strPtr("TECHNOLOGY"),
			WebsiteUrl:       strPtr("https://example.com"),
			Street:           strPtr("456 Test Ave"),
			City:             strPtr("Miami"),
			State:            strPtr("FL"),
			PostalCode:       strPtr("33101"),
			ContactFirstName: strPtr(firstName1),
			ContactLastName:  strPtr(lastName1),
			ContactEmail:     strPtr(email1),
			ContactPhone:     strPtr(phone1),
		})
		if err != nil {
			return err
		}
		campaignBrandID = resp.ID
		if campaignBrandID == 0 {
			return fmt.Errorf("brandId is 0 after create")
		}
		return nil
	})

	// 38. Campaigns.Create
	run("38 Campaigns.create", func() error {
		if campaignBrandID == 0 {
			return fmt.Errorf("no brandId from test 37")
		}
		resp, err := client.Campaigns.Create(campaigns.CampaignRequest{
			BrandID:          campaignBrandID,
			UseCase:          "MARKETING",
			Description:      "Integration test campaign for automated testing",
			MessageFlow:      "Customers opt-in via website form at https://example.com/sms-signup",
			HasEmbeddedLinks: boolPtr(false),
			HasEmbeddedPhone: boolPtr(false),
			IsAgeGated:       boolPtr(false),
			IsDirectLending:  boolPtr(false),
			OptInKeywords:    []string{"START", "YES"},
			OptInMessage:     "You have opted in to receive messages. Reply STOP to unsubscribe.",
			OptInProofUrl:    "https://example.com/opt-in-proof",
			HelpKeywords:     []string{"HELP", "INFO"},
			HelpMessage:      "For help reply HELP or call 1-800-555-0000.",
			OptOutKeywords:   []string{"STOP", "END"},
			OptOutMessage:    "You have been unsubscribed. Reply START to opt back in. STOP",
			SampleMessages: []string{
				"Hello, this is a test message. Reply STOP to unsubscribe.",
				"Reminder: your appointment is tomorrow. Reply HELP for assistance.",
			},
		})
		if err != nil {
			return err
		}
		createdCampaignID = resp.ID
		if createdCampaignID == 0 {
			return fmt.Errorf("campaignId is 0 after create")
		}
		return nil
	})

	// 39. Campaigns.Get
	run("39 Campaigns.get", func() error {
		if createdCampaignID == 0 {
			return fmt.Errorf("no campaignId from test 38")
		}
		_, err := client.Campaigns.Get(createdCampaignID)
		return err
	})

	// 40. Campaigns.List
	run("40 Campaigns.list", func() error {
		_, err := client.Campaigns.List()
		return err
	})

	// 41. Campaigns.Update
	run("41 Campaigns.update", func() error {
		if createdCampaignID == 0 {
			return fmt.Errorf("no campaignId from test 38")
		}
		_, err := client.Campaigns.Update(createdCampaignID, campaigns.CampaignRequest{
			Description: "Updated integration test campaign description",
		})
		return err
	})

	// 42. Campaigns.Delete (also cleans up campaign brand)
	run("42 Campaigns.delete", func() error {
		if createdCampaignID == 0 {
			return fmt.Errorf("no campaignId from test 38")
		}
		if err := client.Campaigns.Delete(createdCampaignID); err != nil {
			return err
		}
		if campaignBrandID != 0 {
			_ = client.Brands.Delete(campaignBrandID)
		}
		return nil
	})

	// ─────────────────────────────────────────────────────────────────────────
	// RESULTS
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Printf("\n==============================================\n")
	fmt.Printf("  RESULTS: %d passed, %d failed\n", passed, failed)
	fmt.Printf("==============================================\n")

	// Print JSON summary for CI parsing
	summary, _ := json.Marshal(map[string]interface{}{
		"sdk":    "go",
		"passed": passed,
		"failed": failed,
		"total":  passed + failed,
	})
	fmt.Printf("\nSUMMARY_JSON: %s\n", string(summary))

	if failed > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
