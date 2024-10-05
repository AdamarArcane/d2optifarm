package main

type ProfileResponse struct {
	Response struct {
		ProfileInventory struct {
			Data struct {
				Items []struct {
					ItemHash       uint32 `json:"itemHash"`
					ItemInstanceID string `json:"itemInstanceId"`
					Quantity       int    `json:"quantity"`
				} `json:"items"`
			} `json:"data"`
		} `json:"profileInventory"`
		// Include other components as needed
	} `json:"Response"`
	ErrorCode   int               `json:"ErrorCode"`
	ErrorStatus string            `json:"ErrorStatus"`
	Message     string            `json:"Message"`
	MessageData map[string]string `json:"MessageData"`
}

type ItemDefinition struct {
	DisplayProperties struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"displayProperties"`
	ItemTypeDisplayName string `json:"itemTypeDisplayName"`
	// Include other fields as needed
}
