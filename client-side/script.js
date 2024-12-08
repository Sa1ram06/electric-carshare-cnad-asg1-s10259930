 // Function to show message
 function showMessage(message, type) {
    const messageDiv = document.getElementById('message') || document.getElementById('rental-message');
    messageDiv.textContent = message;
    messageDiv.className = 'message ' + (type === 'success' ? 'success' : 'error');
    messageDiv.style.display = 'block';

    // Hide message after 3 seconds
    setTimeout(() => {
        messageDiv.style.display = 'none';
    }, 3000);
}