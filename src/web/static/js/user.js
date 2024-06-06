// Save this within ./web/static/js/user.js

$(document).ready(function () {
  $(".btn-logout").click(function (e) {
    Cookies.remove("auth-session");
  });
});
