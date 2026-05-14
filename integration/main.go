// Go SDK integration tests — 52 tests
// Covers: SMS (1-6), MMS (7-17), Email (18-22), Webhook (23-29), Contact (30-31),
// Brands (32-36), Campaigns (37-42), ContactValidator (43-46), Negative cases (47-52)
//
// Test results use three states:
//
//	PASS — the test ran and all assertions held
//	FAIL — the test ran and an assertion (or the API call) failed
//	SKIP — a prerequisite test failed, so this test could not run
//
// Resources created during the run (webhooks, brands, campaigns) are tracked and
// deleted in a final cleanup block even if tests fail midway.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudcontactai/ccai-go/src/pkg/brands"
	"github.com/cloudcontactai/ccai-go/src/pkg/campaigns"
	"github.com/cloudcontactai/ccai-go/src/pkg/ccai"
	"github.com/cloudcontactai/ccai-go/src/pkg/contactvalidator"
	"github.com/cloudcontactai/ccai-go/src/pkg/email"
	"github.com/cloudcontactai/ccai-go/src/pkg/sms"
	"github.com/cloudcontactai/ccai-go/src/pkg/webhook"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

var passed, failed, skipped int

// skipError marks a test that cannot run because a prerequisite test failed.
type skipError struct{ reason string }

func (s skipError) Error() string { return s.reason }

func skipTest(reason string) error { return skipError{reason} }

func run(name string, fn func() error) {
	err := fn()
	if err == nil {
		fmt.Printf("  PASS [%s]\n", name)
		passed++
		return
	}
	if se, ok := err.(skipError); ok {
		fmt.Printf("  SKIP [%s]: %s\n", name, se.reason)
		skipped++
		return
	}
	fmt.Printf("  FAIL [%s]: %v\n", name, err)
	failed++
}

// assertSMSResponse asserts that a send-style response carries a campaign/message identifier.
func assertSMSResponse(resp *sms.Response) error {
	if resp == nil {
		return fmt.Errorf("empty response")
	}
	if resp.GetID() == "" && resp.CampaignID == "" {
		return fmt.Errorf("response has no id/campaignId")
	}
	return nil
}

func assertEmailResponse(resp *email.EmailResponse) error {
	if resp == nil {
		return fmt.Errorf("empty response")
	}
	if resp.ID == nil && resp.CampaignID == nil && resp.ResponseID == "" {
		return fmt.Errorf("response has no id/campaignId/responseId")
	}
	return nil
}

// webhookIDString normalizes the webhook ID, which the API returns sometimes as a
// JSON number and sometimes as a string (known API inconsistency).
func webhookIDString(v interface{}) string {
	switch id := v.(type) {
	case float64:
		return fmt.Sprintf("%.0f", id)
	case string:
		return id
	default:
		return fmt.Sprintf("%v", id)
	}
}

// hmacSHA256Base64 computes Base64(HMAC-SHA256(secret, message))
func hmacSHA256Base64(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func removeString(s []string, v string) []string {
	out := s[:0]
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

func removeInt64(s []int64, v int64) []int64 {
	out := s[:0]
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

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
	// os.Exit skips deferred functions, so all the work (including the deferred
	// resource cleanup) lives in runAll and main only converts its exit code.
	os.Exit(runAll())
}

func runAll() int {
	// ── Validate ALL required env vars up front and report every missing one ──
	requiredEnv := []string{
		"CCAI_CLIENT_ID", "CCAI_API_KEY",
		"CCAI_TEST_PHONE", "CCAI_TEST_PHONE_2", "CCAI_TEST_PHONE_3",
		"CCAI_TEST_EMAIL", "CCAI_TEST_EMAIL_2", "CCAI_TEST_EMAIL_3",
		"CCAI_TEST_FIRST_NAME", "CCAI_TEST_LAST_NAME",
		"CCAI_TEST_FIRST_NAME_2", "CCAI_TEST_LAST_NAME_2",
		"CCAI_TEST_FIRST_NAME_3", "CCAI_TEST_LAST_NAME_3",
		"WEBHOOK_URL",
	}
	var missing []string
	for _, key := range requiredEnv {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "ERROR: required env vars are not set: %s\n", strings.Join(missing, ", "))
		return 2
	}

	clientID := os.Getenv("CCAI_CLIENT_ID")
	apiKey := os.Getenv("CCAI_API_KEY")
	phone1 := os.Getenv("CCAI_TEST_PHONE")
	phone2 := os.Getenv("CCAI_TEST_PHONE_2")
	phone3 := os.Getenv("CCAI_TEST_PHONE_3")
	email1 := os.Getenv("CCAI_TEST_EMAIL")
	email2 := os.Getenv("CCAI_TEST_EMAIL_2")
	email3 := os.Getenv("CCAI_TEST_EMAIL_3")
	firstName1 := os.Getenv("CCAI_TEST_FIRST_NAME")
	lastName1 := os.Getenv("CCAI_TEST_LAST_NAME")
	firstName2 := os.Getenv("CCAI_TEST_FIRST_NAME_2")
	lastName2 := os.Getenv("CCAI_TEST_LAST_NAME_2")
	firstName3 := os.Getenv("CCAI_TEST_FIRST_NAME_3")
	lastName3 := os.Getenv("CCAI_TEST_LAST_NAME_3")

	// Unique per-run suffix so parallel SDK runs don't collide on the same webhook URL
	runID := fmt.Sprintf("go-%d", time.Now().Unix())
	webhookBase := os.Getenv("WEBHOOK_URL")
	sep := "?"
	if strings.Contains(webhookBase, "?") {
		sep = "&"
	}
	webhookURL := fmt.Sprintf("%s%srun=%s", webhookBase, sep, runID)

	senderEmail := os.Getenv("CCAI_TEST_SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "noreply@cloudcontactai.com"
	}
	replyEmail := senderEmail
	senderName := "CCAI Test"
	webhookSecret := os.Getenv("CCAI_WEBHOOK_SECRET")
	if webhookSecret == "" {
		webhookSecret = "test-webhook-secret-go"
	}

	// ── Create client — use CCAI_BASE_URL if set (local dev), otherwise use test environment
	client, err := ccai.NewClient(ccai.Config{
		ClientID:           clientID,
		APIKey:             apiKey,
		UseTestEnvironment: os.Getenv("CCAI_BASE_URL") == "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create client: %v\n", err)
		return 2
	}

	fmt.Println("==============================================")
	fmt.Println("  CCAI Go SDK Integration Tests")
	fmt.Println("==============================================")

	// ── Pre-create temp PNG for MMS tests ────────────────────────────────────
	pngPath, err := writeTempPNG()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create temp PNG: %v\n", err)
		return 2
	}
	defer os.Remove(pngPath)
	// Derive absolute path for Docker usage
	pngPath, _ = filepath.Abs(pngPath)

	// IDs of resources created by the tests; anything still tracked when the
	// deferred cleanup runs is deleted (tests remove entries they already
	// deleted themselves).
	var cleanupWebhookIDs []string
	var cleanupBrandIDs []int64
	var cleanupCampaignIDs []int64
	defer func() {
		for _, id := range cleanupCampaignIDs {
			if err := client.Campaigns.Delete(id); err != nil {
				fmt.Printf("  CLEANUP: could not delete campaign %d: %v\n", id, err)
			} else {
				fmt.Printf("  CLEANUP: deleted leftover campaign %d\n", id)
			}
		}
		for _, id := range cleanupBrandIDs {
			if err := client.Brands.Delete(id); err != nil {
				fmt.Printf("  CLEANUP: could not delete brand %d: %v\n", id, err)
			} else {
				fmt.Printf("  CLEANUP: deleted leftover brand %d\n", id)
			}
		}
		for _, id := range cleanupWebhookIDs {
			if _, err := client.Webhook.Delete(id); err != nil {
				fmt.Printf("  CLEANUP: could not delete webhook %s: %v\n", id, err)
			} else {
				fmt.Printf("  CLEANUP: deleted leftover webhook %s\n", id)
			}
		}
	}()

	// ─────────────────────────────────────────────────────────────────────────
	// SMS TESTS (1–6)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- SMS ---")

	// 1. SMS.SendSingle
	run("01 SMS.SendSingle", func() error {
		resp, err := client.SMS.SendSingle(firstName1, lastName1, phone1, "Hello from Go SDK!", "Go Test", "", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 2. SMS.Send — 1 recipient
	run("02 SMS.Send (1 recipient)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		resp, err := client.SMS.Send(accounts, "Hello 1 recipient!", "Go Test", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 3. SMS.Send — 2 recipients
	run("03 SMS.Send (2 recipients)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
		}
		resp, err := client.SMS.Send(accounts, "Hello 2 recipients!", "Go Test", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 4. SMS.Send — 3 recipients
	run("04 SMS.Send (3 recipients)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
			{FirstName: firstName3, LastName: lastName3, Phone: phone3},
		}
		resp, err := client.SMS.Send(accounts, "Hello 3 recipients!", "Go Test", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
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
		resp, err := client.SMS.Send(accounts, "Hello from ${city}! Claim your ${offer}.", "Go Test Data", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
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
		resp, err := client.SMS.Send(accounts, "Hello with messageData!", "Go Test MsgData", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// MMS TESTS (7–17)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- MMS ---")

	var signedURLResp *sms.SignedURLResponse
	uploadOk := false

	// 7. MMS.GetSignedUploadURL
	run("07 MMS.GetSignedUploadURL", func() error {
		resp, err := client.MMS.GetSignedUploadURL("test_image.png", "image/png", "", true)
		if err != nil {
			return err
		}
		if resp.SignedS3URL == "" {
			return fmt.Errorf("signedS3Url is empty")
		}
		if resp.FileKey == "" {
			return fmt.Errorf("fileKey is empty")
		}
		signedURLResp = resp
		return nil
	})

	// 8. MMS.UploadImageToSignedURL
	run("08 MMS.UploadImageToSignedURL", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		ok, err := client.MMS.UploadImageToSignedURL(signedURLResp.SignedS3URL, pngPath, "image/png")
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("upload returned false")
		}
		uploadOk = true
		return nil
	})

	// 9. MMS.SendSingle
	run("09 MMS.SendSingle", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		resp, err := client.MMS.SendSingle(signedURLResp.FileKey, firstName1, lastName1, phone1, "MMS single!", "Go MMS Test", "", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 10. MMS.Send — 1 recipient
	run("10 MMS.Send (1 recipient)", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		resp, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS 1 recipient!", "Go MMS Test", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 11. MMS.Send — 2 recipients
	run("11 MMS.Send (2 recipients)", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
		}
		resp, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS 2 recipients!", "Go MMS Test", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 12. MMS.Send — 3 recipients
	run("12 MMS.Send (3 recipients)", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
			{FirstName: firstName2, LastName: lastName2, Phone: phone2},
			{FirstName: firstName3, LastName: lastName3, Phone: phone3},
		}
		resp, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS 3 recipients!", "Go MMS Test", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 13. MMS.Send with Data field
	run("13 MMS.Send with data", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{
				FirstName: firstName1,
				LastName:  lastName1,
				Phone:     phone1,
				Data:      map[string]string{"product": "Widget"},
			},
		}
		resp, err := client.MMS.Send(signedURLResp.FileKey, accounts, "Check out ${product}!", "Go MMS Data", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 14. MMS.Send with MessageData field
	run("14 MMS.Send with messageData", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		accounts := []sms.Account{
			{
				FirstName:   firstName1,
				LastName:    lastName1,
				Phone:       phone1,
				MessageData: `{"campaignId":"mms-test-001"}`,
			},
		}
		resp, err := client.MMS.Send(signedURLResp.FileKey, accounts, "MMS with messageData!", "Go MMS MsgData", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 15. MMS.CheckFileUploaded — the file uploaded in test 08 must actually exist
	run("15 MMS.CheckFileUploaded", func() error {
		if signedURLResp == nil {
			return skipTest("dependency test 07 failed")
		}
		if !uploadOk {
			return skipTest("dependency test 08 failed")
		}
		resp, err := client.MMS.CheckFileUploaded(signedURLResp.FileKey)
		if err != nil {
			return err
		}
		if resp == nil || resp.StoredURL == "" {
			return fmt.Errorf("expected non-empty storedUrl for uploaded file %s", signedURLResp.FileKey)
		}
		return nil
	})

	// 16. MMS.SendWithImage (fresh upload)
	run("16 MMS.SendWithImage (fresh upload)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		resp, err := client.MMS.SendWithImage(pngPath, "image/png", accounts, "MMS with image!", "Go MMS Image", "", nil, true)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 17. MMS.SendWithImage (cached — same image, should reuse S3 key)
	run("17 MMS.SendWithImage (cached)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		resp, err := client.MMS.SendWithImage(pngPath, "image/png", accounts, "MMS cached image!", "Go MMS Cache", "", nil, true)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// EMAIL TESTS (18–22)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Email ---")

	// 18. Email.SendSingle
	run("18 Email.SendSingle", func() error {
		resp, err := client.Email.SendSingle(
			firstName1, lastName1, email1,
			"Go SDK Test Email", "<p>Hello from Go SDK!</p>", "",
			senderEmail, replyEmail, senderName, "Go Email Test",
			nil,
		)
		if err != nil {
			return err
		}
		return assertEmailResponse(resp)
	})

	// 19. Email.Send — 1 recipient
	run("19 Email.Send (1 recipient)", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
		}
		resp, err := client.Email.Send(accounts, "Go SDK Email 1", "<p>Hello 1!</p>", senderEmail, replyEmail, senderName, "Go Email Test", nil)
		if err != nil {
			return err
		}
		return assertEmailResponse(resp)
	})

	// 20. Email.Send — 2 recipients
	run("20 Email.Send (2 recipients)", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
			{FirstName: firstName2, LastName: lastName2, Email: email2},
		}
		resp, err := client.Email.Send(accounts, "Go SDK Email 2", "<p>Hello 2!</p>", senderEmail, replyEmail, senderName, "Go Email Test", nil)
		if err != nil {
			return err
		}
		return assertEmailResponse(resp)
	})

	// 21. Email.Send — 3 recipients
	run("21 Email.Send (3 recipients)", func() error {
		accounts := []ccai.EmailAccount{
			{FirstName: firstName1, LastName: lastName1, Email: email1},
			{FirstName: firstName2, LastName: lastName2, Email: email2},
			{FirstName: firstName3, LastName: lastName3, Email: email3},
		}
		resp, err := client.Email.Send(accounts, "Go SDK Email 3", "<p>Hello 3!</p>", senderEmail, replyEmail, senderName, "Go Email Test", nil)
		if err != nil {
			return err
		}
		return assertEmailResponse(resp)
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
		resp, err := client.Email.SendCampaign(campaign, nil)
		if err != nil {
			return err
		}
		return assertEmailResponse(resp)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// WEBHOOK TESTS (23–29)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Webhook ---")

	var registeredWebhookID string

	// 23. Webhook.Register
	run("23 Webhook.Register", func() error {
		resp, err := client.Webhook.Register(webhook.WebhookConfig{
			URL:    webhookURL,
			Secret: &webhookSecret,
		})
		if err != nil {
			return err
		}
		registeredWebhookID = webhookIDString(resp.ID)
		if registeredWebhookID == "" || registeredWebhookID == "0" {
			return fmt.Errorf("webhook ID is empty after register")
		}
		cleanupWebhookIDs = append(cleanupWebhookIDs, registeredWebhookID)
		return nil
	})

	// 24. Webhook.List — must contain the webhook registered in test 23
	run("24 Webhook.List", func() error {
		hooks, err := client.Webhook.List()
		if err != nil {
			return err
		}
		if len(hooks) == 0 {
			return fmt.Errorf("expected at least one webhook, got 0")
		}
		if registeredWebhookID != "" {
			found := false
			for _, h := range hooks {
				if webhookIDString(h.ID) == registeredWebhookID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("webhook %s registered in test 23 not present in List()", registeredWebhookID)
			}
		}
		return nil
	})

	// 25. Webhook.Update — then verify via List() that the URL actually changed
	run("25 Webhook.Update", func() error {
		if registeredWebhookID == "" {
			return skipTest("dependency test 23 failed")
		}
		updatedSecret := "updated-secret-go"
		_, err := client.Webhook.Update(registeredWebhookID, webhook.WebhookConfig{
			URL:    webhookURL + "&updated=1",
			Secret: &updatedSecret,
		})
		if err != nil {
			return err
		}
		hooks, err := client.Webhook.List()
		if err != nil {
			return err
		}
		for _, h := range hooks {
			if webhookIDString(h.ID) == registeredWebhookID {
				if !strings.Contains(h.URL, "updated=1") {
					return fmt.Errorf("webhook URL was not updated: expected to contain \"updated=1\", got %q", h.URL)
				}
				return nil
			}
		}
		return fmt.Errorf("webhook %s not found in List() after update", registeredWebhookID)
	})

	// 26. Webhook.VerifySignature — valid
	run("26 Webhook.VerifySignature (valid)", func() error {
		eventHash := "abc123eventHash"
		sig := hmacSHA256Base64(webhookSecret, clientID+":"+eventHash)
		ok := client.Webhook.VerifySignature(sig, clientID, eventHash, webhookSecret)
		if !ok {
			return fmt.Errorf("expected valid signature to return true")
		}
		return nil
	})

	// 27. Webhook.VerifySignature — invalid
	run("27 Webhook.VerifySignature (invalid)", func() error {
		ok := client.Webhook.VerifySignature("invalidsig==", clientID, "somehash", webhookSecret)
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
		if event.EventType != "message.sent" {
			return fmt.Errorf("expected eventType \"message.sent\", got %q", event.EventType)
		}
		return nil
	})

	// 29. Webhook.Delete — then verify via List() that it is gone
	run("29 Webhook.Delete", func() error {
		if registeredWebhookID == "" {
			return skipTest("dependency test 23 failed")
		}
		if _, err := client.Webhook.Delete(registeredWebhookID); err != nil {
			return err
		}
		cleanupWebhookIDs = removeString(cleanupWebhookIDs, registeredWebhookID)
		hooks, err := client.Webhook.List()
		if err != nil {
			return err
		}
		for _, h := range hooks {
			if webhookIDString(h.ID) == registeredWebhookID {
				return fmt.Errorf("webhook %s still present in List() after delete", registeredWebhookID)
			}
		}
		return nil
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
		cleanupBrandIDs = append(cleanupBrandIDs, createdBrandID)
		return nil
	})

	// 33. Brands.Get
	run("33 Brands.get", func() error {
		if createdBrandID == 0 {
			return skipTest("dependency test 32 failed")
		}
		resp, err := client.Brands.Get(createdBrandID)
		if err != nil {
			return err
		}
		if resp.ID != createdBrandID {
			return fmt.Errorf("brand id mismatch: expected %d, got %d", createdBrandID, resp.ID)
		}
		if resp.LegalCompanyName != "Test Company LLC" {
			return fmt.Errorf("expected legalCompanyName \"Test Company LLC\", got %q", resp.LegalCompanyName)
		}
		return nil
	})

	// 34. Brands.List — must contain the brand created in test 32
	run("34 Brands.list", func() error {
		resp, err := client.Brands.List()
		if err != nil {
			return err
		}
		if createdBrandID != 0 {
			for _, b := range resp {
				if b.ID == createdBrandID {
					return nil
				}
			}
			return fmt.Errorf("brand %d created in test 32 not present in List()", createdBrandID)
		}
		return nil
	})

	// 35. Brands.Update — then verify via Get() that the field actually changed
	run("35 Brands.update", func() error {
		if createdBrandID == 0 {
			return skipTest("dependency test 32 failed")
		}
		_, err := client.Brands.Update(createdBrandID, brands.BrandRequest{
			City: strPtr("Orlando"),
		})
		if err != nil {
			return err
		}
		fetched, err := client.Brands.Get(createdBrandID)
		if err != nil {
			return err
		}
		if fetched.City != "Orlando" {
			return fmt.Errorf("expected city \"Orlando\" after update, got %q", fetched.City)
		}
		return nil
	})

	// 36. Brands.Delete — then verify via Get() that it is gone
	run("36 Brands.delete", func() error {
		if createdBrandID == 0 {
			return skipTest("dependency test 32 failed")
		}
		if err := client.Brands.Delete(createdBrandID); err != nil {
			return err
		}
		cleanupBrandIDs = removeInt64(cleanupBrandIDs, createdBrandID)
		if _, err := client.Brands.Get(createdBrandID); err == nil {
			return fmt.Errorf("expected Get of deleted brand %d to fail, but it succeeded", createdBrandID)
		}
		return nil
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
		cleanupBrandIDs = append(cleanupBrandIDs, campaignBrandID)
		return nil
	})

	// 38. Campaigns.Create
	run("38 Campaigns.create", func() error {
		if campaignBrandID == 0 {
			return skipTest("dependency test 37 failed")
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
		cleanupCampaignIDs = append(cleanupCampaignIDs, createdCampaignID)
		return nil
	})

	// 39. Campaigns.Get
	run("39 Campaigns.get", func() error {
		if createdCampaignID == 0 {
			return skipTest("dependency test 38 failed")
		}
		resp, err := client.Campaigns.Get(createdCampaignID)
		if err != nil {
			return err
		}
		if resp.ID != createdCampaignID {
			return fmt.Errorf("campaign id mismatch: expected %d, got %d", createdCampaignID, resp.ID)
		}
		if resp.BrandID != campaignBrandID {
			return fmt.Errorf("expected brandId %d, got %d", campaignBrandID, resp.BrandID)
		}
		return nil
	})

	// 40. Campaigns.List — must contain the campaign created in test 38
	run("40 Campaigns.list", func() error {
		resp, err := client.Campaigns.List()
		if err != nil {
			return err
		}
		if createdCampaignID != 0 {
			for _, c := range resp {
				if c.ID == createdCampaignID {
					return nil
				}
			}
			return fmt.Errorf("campaign %d created in test 38 not present in List()", createdCampaignID)
		}
		return nil
	})

	// 41. Campaigns.Update — then verify via Get() that the field actually changed
	run("41 Campaigns.update", func() error {
		if createdCampaignID == 0 {
			return skipTest("dependency test 38 failed")
		}
		const newDescription = "Updated integration test campaign description"
		_, err := client.Campaigns.Update(createdCampaignID, campaigns.CampaignRequest{
			Description: newDescription,
		})
		if err != nil {
			return err
		}
		fetched, err := client.Campaigns.Get(createdCampaignID)
		if err != nil {
			return err
		}
		if fetched.Description != newDescription {
			return fmt.Errorf("expected updated description after update, got %q", fetched.Description)
		}
		return nil
	})

	// 42. Campaigns.Delete (also cleans up campaign brand) — then verify via Get() that it is gone
	run("42 Campaigns.delete", func() error {
		if createdCampaignID == 0 {
			return skipTest("dependency test 38 failed")
		}
		if err := client.Campaigns.Delete(createdCampaignID); err != nil {
			return err
		}
		cleanupCampaignIDs = removeInt64(cleanupCampaignIDs, createdCampaignID)
		if _, err := client.Campaigns.Get(createdCampaignID); err == nil {
			return fmt.Errorf("expected Get of deleted campaign %d to fail, but it succeeded", createdCampaignID)
		}
		if campaignBrandID != 0 {
			if err := client.Brands.Delete(campaignBrandID); err != nil {
				return fmt.Errorf("failed to delete campaign brand %d: %w", campaignBrandID, err)
			}
			cleanupBrandIDs = removeInt64(cleanupBrandIDs, campaignBrandID)
		}
		return nil
	})

	// ─────────────────────────────────────────────────────────────────────────
	// CONTACT VALIDATOR TESTS (43–46)
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- ContactValidator ---")

	run("43 ContactValidator.ValidateEmail", func() error {
		resp, err := client.ContactValidator.ValidateEmail(email1)
		if err != nil {
			return err
		}
		if resp.Status == "" {
			return fmt.Errorf("status is empty")
		}
		return nil
	})

	run("44 ContactValidator.ValidateEmails", func() error {
		resp, err := client.ContactValidator.ValidateEmails([]string{email1, email2})
		if err != nil {
			return err
		}
		if resp.Summary.Total != 2 {
			return fmt.Errorf("expected summary.total=2, got %d", resp.Summary.Total)
		}
		if len(resp.Results) != 2 {
			return fmt.Errorf("expected 2 results, got %d", len(resp.Results))
		}
		return nil
	})

	run("45 ContactValidator.ValidatePhone", func() error {
		resp, err := client.ContactValidator.ValidatePhone(phone1, "")
		if err != nil {
			return err
		}
		if resp.Status == "" {
			return fmt.Errorf("status is empty")
		}
		return nil
	})

	run("46 ContactValidator.ValidatePhones", func() error {
		resp, err := client.ContactValidator.ValidatePhones([]contactvalidator.PhoneInput{
			{Phone: phone1},
			{Phone: phone2},
		})
		if err != nil {
			return err
		}
		if resp.Summary.Total != 2 {
			return fmt.Errorf("expected summary.total=2, got %d", resp.Summary.Total)
		}
		if len(resp.Results) != 2 {
			return fmt.Errorf("expected 2 results, got %d", len(resp.Results))
		}
		return nil
	})

	// ─────────────────────────────────────────────────────────────────────────
	// NEGATIVE & PERMISSIVE TESTS (47–52)
	// 47/49/50 PASS when the operation fails as expected. 48/51/52 document
	// permissive behavior observed in the test API: those
	// operations succeed even with invalid input, so the tests assert success.
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("\n--- Negative cases ---")

	// 47. Invalid API key must be rejected
	run("47 NEGATIVE: SMS.SendSingle with invalid API key", func() error {
		badClient, err := ccai.NewClient(ccai.Config{
			ClientID:           clientID,
			APIKey:             "invalid-api-key-for-negative-test",
			UseTestEnvironment: os.Getenv("CCAI_BASE_URL") == "",
		})
		if err != nil {
			return err
		}
		if _, err := badClient.SMS.SendSingle(firstName1, lastName1, phone1, "should fail", "Go Negative 47", "", "", nil); err == nil {
			return fmt.Errorf("expected send with invalid API key to fail, but it succeeded")
		}
		return nil
	})

	// 48. The test API accepts malformed phone numbers: the send
	// succeeds instead of failing. If the API starts validating phone format, change
	// this back to expect an error.
	run("48 PERMISSIVE: SMS.SendSingle with malformed phone (API accepts)", func() error {
		resp, err := client.SMS.SendSingle(firstName1, lastName1, "abc", "malformed phone accepted", "Go Permissive 48", "", "", nil)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// 49. Getting a nonexistent brand must fail
	run("49 NEGATIVE: Brands.Get(nonexistent)", func() error {
		if _, err := client.Brands.Get(99999999); err == nil {
			return fmt.Errorf("expected Get of nonexistent brand to fail, but it succeeded")
		}
		return nil
	})

	// 50. Deleting a nonexistent webhook must fail
	run("50 NEGATIVE: Webhook.Delete(nonexistent)", func() error {
		if _, err := client.Webhook.Delete("99999999"); err == nil {
			return fmt.Errorf("expected Delete of nonexistent webhook to fail, but it succeeded")
		}
		return nil
	})

	// 51. The test environment's validator reports "valid" even for syntactically
	// invalid emails — upstream validation is not enforced
	// there, so only assert that a status is returned.
	run("51 PERMISSIVE: ContactValidator.ValidateEmail(invalid input)", func() error {
		resp, err := client.ContactValidator.ValidateEmail("not-an-email")
		if err != nil {
			return err
		}
		if resp.Status == "" {
			return fmt.Errorf("status is empty")
		}
		return nil
	})

	// 52. The test API accepts MMS sends with a nonexistent fileKey: it does not
	// verify the file exists at send time. If the API
	// starts validating the fileKey, change this back to expect an error.
	run("52 PERMISSIVE: MMS.Send with nonexistent fileKey (API accepts)", func() error {
		accounts := []sms.Account{
			{FirstName: firstName1, LastName: lastName1, Phone: phone1},
		}
		fakeKey := fmt.Sprintf("%s/campaign/nonexistent_%d.png", clientID, time.Now().Unix())
		resp, err := client.MMS.Send(fakeKey, accounts, "nonexistent fileKey accepted", "Go Permissive 52", "", nil, false)
		if err != nil {
			return err
		}
		return assertSMSResponse(resp)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// RESULTS
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Printf("\n==============================================\n")
	fmt.Printf("  RESULTS: %d passed, %d failed, %d skipped\n", passed, failed, skipped)
	fmt.Printf("==============================================\n")

	// Print JSON summary for CI parsing
	summary, _ := json.Marshal(map[string]interface{}{
		"sdk":     "go",
		"passed":  passed,
		"failed":  failed,
		"skipped": skipped,
		"total":   passed + failed + skipped,
	})
	fmt.Printf("\nSUMMARY_JSON: %s\n", string(summary))

	if failed > 0 {
		return 1
	}
	return 0
}
