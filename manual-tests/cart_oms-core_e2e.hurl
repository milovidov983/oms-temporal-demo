# Create a new order
POST http://localhost:9999/api/cart
Content-Type: application/json
{
    "items": [
        {
            "product_id": "product1",
            "price": 5.0,
            "quantity": 1
        }
    ]
}

# Verify the response status code
HTTP 200
# Capture the order_id from response
[Captures]
order_id: jsonpath "$.order_id"

# Assert that we received an order_id
[Asserts]
jsonpath "$.order_id" exists

# Check the order status using captured order_id
GET http://localhost:9999/api/cart/status?order_id={{order_id}}

# Verify the response status code
HTTP 200
[Asserts]
jsonpath "$.text" == "Order status: CREATED"