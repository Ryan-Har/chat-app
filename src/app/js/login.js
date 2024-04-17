console.log("login.js loaded")

document.getElementById("loginButton").addEventListener('click', function() {
    console.log("login button clicked");
    logIn();
});


function logIn() {
    let emailElement = document.getElementById("username")
    let passwordElement = document.getElementById("password")
    fetch("/handlelogin", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            email: emailElement.value,
            password: passwordElement.value
        })
    })
    .then((response) => {
        console.log("response:", response);
        return response.json();
    })
    .then((data) => {
        const result = parseInt(data);
        if (result == 0) {
            alert("Login failed. Please try again.");
            return;
        }
        let parentwindow = window.parent;
        let dataToSend = {operation: "Login", message: result};
        parentwindow.postMessage(dataToSend, "*");
    });
}