document.addEventListener('DOMContentLoaded', function() {
    const usersTableBody = document.getElementById('usersTableBody');
    const searchInput = document.getElementById('searchInput');
    const notification = document.getElementById('notification');

    function loadUsers() {
        return JSON.parse(localStorage.getItem('users')) || [];
    }

    
    function showNotification(message, type) {
        notification.className = `alert alert-${type}`;
        notification.textContent = message;
        notification.classList.remove('d-none');
        setTimeout(() => notification.classList.add('d-none'), 3000);
    }

    
    searchInput.addEventListener('input', displayUsers);

    function displayUsers() {
        const searchQuery = searchInput.value.toLowerCase();
        usersTableBody.innerHTML = '';
        const users = loadUsers();

        users
            .filter(user => user.name.toLowerCase().includes(searchQuery) || user.email.toLowerCase().includes(searchQuery))
            .forEach((user, index) => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${user.name}</td>
                    <td>${user.surname}</td>
                    <td>${user.email}</td>
                    <td>${user.role}</td>
                    <td>
                        <button class="btn btn-warning btn-sm" onclick="editUser(${index})">Edit</button>
                        <button class="btn btn-danger btn-sm" onclick="deleteUser(${index})">Delete</button>
                    </td>
                `;
                usersTableBody.appendChild(row);
            });
    }

    
    window.deleteUser = function(index) {
        const users = loadUsers();
        const confirmDelete = confirm(`Are you sure you want to delete user ${users[index].name}?`);
        if (confirmDelete) {
            users.splice(index, 1);
            localStorage.setItem('users', JSON.stringify(users));
            displayUsers();
            showNotification('User deleted successfully', 'success');
        }
    }

    
    window.editUser = function(index) {
        const users = loadUsers();
        const user = users[index];
        
        const newName = prompt("Enter new name:", user.name);
        const newSurname = prompt("Enter new surname:", user.surname);
        const newEmail = prompt("Enter new email:", user.email);

        if (newName && newSurname && newEmail) {
            users[index] = { ...user, name: newName, surname: newSurname, email: newEmail };
            localStorage.setItem('users', JSON.stringify(users));
            displayUsers();
            showNotification('User updated successfully', 'success');
        }
    }

    
    window.sortTable = function(field) {
        const users = loadUsers();
        users.sort((a, b) => a[field].localeCompare(b[field]));
        localStorage.setItem('users', JSON.stringify(users));
        displayUsers();
        showNotification(`Sorted by ${field}`, 'info');
    }

    displayUsers();
});
