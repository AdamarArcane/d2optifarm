package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GetMembershipData(client *http.Client) (string, int, error) {
	resp, err := client.Get("https://www.bungie.net/Platform/User/GetMembershipsForCurrentUser/")
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Response struct {
			DestinyMemberships []struct {
				MembershipID   string `json:"membershipId"`
				MembershipType int    `json:"membershipType"`
				DisplayName    string `json:"displayName"`
			} `json:"destinyMemberships"`
		} `json:"Response"`
		ErrorCode   int               `json:"ErrorCode"`
		ErrorStatus string            `json:"ErrorStatus"`
		Message     string            `json:"Message"`
		MessageData map[string]string `json:"MessageData"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", 0, err
	}

	if result.ErrorCode != 1 {
		return "", 0, fmt.Errorf("API error: %s", result.Message)
	}

	if len(result.Response.DestinyMemberships) == 0 {
		return "", 0, fmt.Errorf("no Destiny memberships found")
	}

	// Use the first membership
	membershipID := result.Response.DestinyMemberships[0].MembershipID
	membershipType := result.Response.DestinyMemberships[0].MembershipType
	return membershipID, membershipType, nil
}

func GetProfileData(client *http.Client, membershipID string, membershipType int) (*ProfileResponse, error) {
	// Define the components you want to retrieve
	components := []int{100, 102, 200, 201, 205, 300, 305, 310, 800}

	// Build the URL with query parameters
	url := fmt.Sprintf("https://www.bungie.net/Platform/Destiny2/%d/Profile/%s/?components=%s",
		membershipType, membershipID, strings.Join(IntSliceToStringSlice(components), ","))

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.ErrorCode != 1 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func GetItemDefinition(client *http.Client, itemHash uint32) (*ItemDefinition, error) {
	url := fmt.Sprintf("https://www.bungie.net/Platform/Destiny2/Manifest/DestinyInventoryItemDefinition/%d/", itemHash)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Response    ItemDefinition    `json:"Response"`
		ErrorCode   int               `json:"ErrorCode"`
		ErrorStatus string            `json:"ErrorStatus"`
		Message     string            `json:"Message"`
		MessageData map[string]string `json:"MessageData"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.ErrorCode != 1 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result.Response, nil
}
