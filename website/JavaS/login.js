document.querySelector('#loginForm').addEventListener('submit', function(event) {
    event.preventDefault();

    const email = document.querySelector('#email').value;
    const password = document.querySelector('#password').value;
    const users = JSON.parse(localStorage.getItem('users')) || [];

    
    const user = users.find(u => u.email === email && u.password === password);

    if (user) {
        alert('Login successful!');
        
        localStorage.setItem('currentUser', JSON.stringify(user));

        if (user.role === 'admin') {
            window.location.href = 'admin_panel.html'; 
        } else {
            window.location.href = 'index.html'; 
        }
    } else {
        alert('Invalid email or password.');
    }
});
