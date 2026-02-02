package challeng13

import (
	"database/sql"
	"os"
	"testing"
)

const testDBPath = "test_inventory.db"

func setupTestDB(t *testing.T) *sql.DB {
	// Remove any existing test database
	os.Remove(testDBPath)

	// Initialize a new test database
	db, err := InitDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	return db
}

func cleanupTestDB() {
	os.Remove(testDBPath)
}

func TestCreateProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	testCases := []struct {
		name    string
		product Product
	}{
		{
			name: "Create Basic Product",
			product: Product{
				Name:     "Test Product",
				Price:    9.99,
				Quantity: 100,
				Category: "Test",
			},
		},
		{
			name: "Create Zero Price Product",
			product: Product{
				Name:     "Free Product",
				Price:    0,
				Quantity: 10,
				Category: "Free",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			product := tc.product

			err := store.CreateProduct(&product)
			if err != nil {
				t.Fatalf("Failed to create product: %v", err)
			}

			if product.ID <= 0 {
				t.Errorf("Expected product ID to be set after creation, got %d", product.ID)
			}

			// Verify product was created by retrieving it
			retrieved, err := store.GetProduct(product.ID)
			if err != nil {
				t.Fatalf("Failed to retrieve created product: %v", err)
			}

			if retrieved.Name != product.Name {
				t.Errorf("Expected name %s, got %s", product.Name, retrieved.Name)
			}

			if retrieved.Price != product.Price {
				t.Errorf("Expected price %f, got %f", product.Price, retrieved.Price)
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	// Create a product to retrieve
	product := &Product{
		Name:     "Test Product",
		Price:    9.99,
		Quantity: 100,
		Category: "Test",
	}
	err := store.CreateProduct(product)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	testCases := []struct {
		name        string
		id          int64
		expectError bool
	}{
		{
			name:        "Get Existing Product",
			id:          product.ID,
			expectError: false,
		},
		{
			name:        "Get Non-Existent Product",
			id:          product.ID + 1000, // ID that doesn't exist
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retrieved, err := store.GetProduct(tc.id)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error retrieving non-existent product, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Failed to retrieve product: %v", err)
				}

				if retrieved.ID != product.ID {
					t.Errorf("Expected ID %d, got %d", product.ID, retrieved.ID)
				}

				if retrieved.Name != product.Name {
					t.Errorf("Expected name %s, got %s", product.Name, retrieved.Name)
				}
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	// Create a product to update
	product := &Product{
		Name:     "Original Name",
		Price:    9.99,
		Quantity: 100,
		Category: "Test",
	}
	err := store.CreateProduct(product)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Update the product
	product.Name = "Updated Name"
	product.Price = 19.99
	product.Quantity = 50

	err = store.UpdateProduct(product)
	if err != nil {
		t.Fatalf("Failed to update product: %v", err)
	}

	// Verify the update
	updated, err := store.GetProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated product: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Update failed: expected name 'Updated Name', got '%s'", updated.Name)
	}

	if updated.Price != 19.99 {
		t.Errorf("Update failed: expected price 19.99, got %f", updated.Price)
	}

	if updated.Quantity != 50 {
		t.Errorf("Update failed: expected quantity 50, got %d", updated.Quantity)
	}
}

func TestDeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	// Create a product to delete
	product := &Product{
		Name:     "To Be Deleted",
		Price:    9.99,
		Quantity: 100,
		Category: "Test",
	}
	err := store.CreateProduct(product)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Delete the product
	err = store.DeleteProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	// Verify the deletion
	_, err = store.GetProduct(product.ID)
	if err == nil {
		t.Errorf("Expected error when retrieving deleted product, got nil")
	}
}

func TestListProducts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	// Create products with different categories
	productsToCreate := []Product{
		{Name: "Product 1", Price: 9.99, Quantity: 10, Category: "Electronics"},
		{Name: "Product 2", Price: 19.99, Quantity: 20, Category: "Electronics"},
		{Name: "Product 3", Price: 29.99, Quantity: 30, Category: "Books"},
		{Name: "Product 4", Price: 39.99, Quantity: 40, Category: "Clothing"},
	}

	for i := range productsToCreate {
		err := store.CreateProduct(&productsToCreate[i])
		if err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	testCases := []struct {
		name         string
		category     string
		expectedSize int
	}{
		{
			name:         "List All Products",
			category:     "",
			expectedSize: 4,
		},
		{
			name:         "List Electronics Products",
			category:     "Electronics",
			expectedSize: 2,
		},
		{
			name:         "List Books Products",
			category:     "Books",
			expectedSize: 1,
		},
		{
			name:         "List Non-Existent Category",
			category:     "NonExistent",
			expectedSize: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			products, err := store.ListProducts(tc.category)
			if err != nil {
				t.Fatalf("Failed to list products: %v", err)
			}

			if len(products) != tc.expectedSize {
				t.Errorf("Expected %d products, got %d", tc.expectedSize, len(products))
			}

			if tc.category != "" {
				for _, p := range products {
					if p.Category != tc.category {
						t.Errorf("Expected category %s, got %s", tc.category, p.Category)
					}
				}
			}
		})
	}
}

func TestBatchUpdateInventory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	// Create products
	products := []Product{
		{Name: "Product 1", Price: 9.99, Quantity: 10, Category: "Test"},
		{Name: "Product 2", Price: 19.99, Quantity: 20, Category: "Test"},
	}

	for i := range products {
		err := store.CreateProduct(&products[i])
		if err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	// Prepare batch updates
	updates := map[int64]int{
		products[0].ID: 5,  // Reduce by 5
		products[1].ID: 15, // Reduce by 5
	}

	// Perform batch update
	err := store.BatchUpdateInventory(updates)
	if err != nil {
		t.Fatalf("Failed to perform batch update: %v", err)
	}

	// Verify updates
	p1, err := store.GetProduct(products[0].ID)
	if err != nil {
		t.Fatalf("Failed to retrieve product 1: %v", err)
	}
	if p1.Quantity != 5 {
		t.Errorf("Expected quantity 5 for product 1, got %d", p1.Quantity)
	}

	p2, err := store.GetProduct(products[1].ID)
	if err != nil {
		t.Fatalf("Failed to retrieve product 2: %v", err)
	}
	if p2.Quantity != 15 {
		t.Errorf("Expected quantity 15 for product 2, got %d", p2.Quantity)
	}
}

func TestBatchUpdateInventoryRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestDB()

	store := NewProductStore(db)

	// Create one product
	product := &Product{
		Name:     "Product 1",
		Price:    9.99,
		Quantity: 10,
		Category: "Test",
	}
	err := store.CreateProduct(product)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Try to update with a non-existent product ID
	updates := map[int64]int{
		product.ID:      5,
		product.ID + 99: 15, // This product doesn't exist
	}

	// This should fail and roll back
	err = store.BatchUpdateInventory(updates)
	if err == nil {
		t.Fatalf("Expected error for non-existent product, got nil")
	}

	// Verify the first product was not updated (rollback worked)
	p1, err := store.GetProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve product: %v", err)
	}
	if p1.Quantity != 10 {
		t.Errorf("Expected quantity to remain 10 after rollback, got %d", p1.Quantity)
	}
}
