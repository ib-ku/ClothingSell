function regButton() {
    alert('You successfully registered!');
}


function addToCart(name, price, image) {
    const product = {
        name: name,
        price: price,
        image: image
    };

    const cart = JSON.parse(localStorage.getItem('cart')) || [];
    cart.push(product);
    localStorage.setItem('cart', JSON.stringify(cart));

    alert(`${name} added to cart`);
}




