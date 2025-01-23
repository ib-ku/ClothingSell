document.querySelector('#registrationForm').addEventListener('submit', function(event) {
    event.preventDefault();

    const name = document.querySelector('#name');
    const surname = document.querySelector('#surname');
    const email = document.querySelector('#email');
    const password = document.querySelector('#password');

    let isValid = true;

    const doesNotStartWithDigit = (str) => !/^\d/.test(str);

    if (!doesNotStartWithDigit(name.value)) {
        name.classList.add('input-error');
        isValid = false;
        alert("name should not start with digit")
    } else {
        name.classList.remove('input-error');
    }

    if (!doesNotStartWithDigit(surname.value)) {
        surname.classList.add('input-error');
        isValid = false;
        alert("surname should not start with digit")
    } else {
        surname.classList.remove('input-error');
    }

    if (!doesNotStartWithDigit(email.value)) {
        email.classList.add('input-error');
        isValid = false;
        alert("email should not start with digit")
    } else {
        email.classList.remove('input-error');
    }

    const passwordRegex = /^(?=.*[A-Z])(?=.*\d).{8,}$/;
    if (!passwordRegex.test(password.value)) {
        password.classList.add('input-error');
        isValid = false;
        alert("Must be at least 8 characters long. Must contain at least one uppercase letter. Must contain at least one digit. Should not start with a digit")
    } else {
        password.classList.remove('input-error');
    }

    if (!isValid) {
        alert('fill out the form correctly');
        return;
    }

    const users = JSON.parse(localStorage.getItem('users')) || [];

    const newUser = {
        name: name.value,
        surname: surname.value,
        email: email.value,
        password: password.value,
        role: "user"
    };

    users.push(newUser);
    localStorage.setItem('users', JSON.stringify(users));

    alert('Registration successful!');
    window.location.href = 'login.html';
});