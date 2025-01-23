document.addEventListener('DOMContentLoaded', displayCart);

function displayCart() {
    const cart = JSON.parse(localStorage.getItem('cart')) || [];
    const cartItems = document.getElementById('cartItems');
    const totalPriceElement = document.getElementById('totalPrice'); 
    let totalPrice = 0;

    cartItems.innerHTML = '';  

    cart.forEach((item, index) => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><img src="${item.image}" alt="${item.name}" width="50"> ${item.name}</td>
            <td>${item.price} TG</td>
            <td>
                <button class="btn btn-success btn-sm" onclick="purchaseItem(${index})">Buy</button>
                <button class="btn btn-danger btn-sm" onclick="removeFromCart(${index})">Remove</button>
            </td>
        `;
        cartItems.appendChild(row);

        
        totalPrice += item.price;
    });

    
    totalPriceElement.textContent = `Total Price: ${totalPrice} TG`;
}

function purchaseItem(index) {
    const cart = JSON.parse(localStorage.getItem('cart')) || [];
    const item = cart[index];

    if (confirm(`Confirm purchase of ${item.name} for ${item.price} TG?`)) {
        cart.splice(index, 1);
        localStorage.setItem('cart', JSON.stringify(cart));
        displayCart();  
        alert(`${item.name} purchased successfully!`);
    }
}

function purchaseAllItems() {
    const cart = JSON.parse(localStorage.getItem('cart')) || [];
    const totalPrice = cart.reduce((total, item) => total + item.price, 0);

    if (cart.length > 0 && confirm(`Confirm purchase of all items for ${totalPrice} TG?`)) {
        localStorage.removeItem('cart'); 
        displayCart();
        alert('All items purchased successfully!');
    } else if (cart.length === 0) {
        alert('Cart is empty.');
    }
}


function removeFromCart(index) {
    const cart = JSON.parse(localStorage.getItem('cart')) || [];
    cart.splice(index, 1); 
    localStorage.setItem('cart', JSON.stringify(cart));
    displayCart(); 
}

function clearCart() {
    localStorage.removeItem('cart');
    displayCart();
    alert('Cart clear');
}
