$(document).ready(function() {
    // process the form
    $('form').submit(function(event) {
        // get the form data
        // there are many ways to get this data using jQuery (you can use the class or id also)
        var divId = $(this).find('input[name=game_id]').val()
        var formData = {
            'username'              : $(this).find('input[name=username]').val(),
            'content'             : $(this).find('input[name=content]').val(),
            'game_id'               :$(this).find('input[name=game_id]').val(),
        };
        // process the form
        $.ajax({
            type        : 'POST', // define the type of HTTP verb we want to use (POST for our form)
            url         : '/postComment', // the url where we want to POST
            data        : JSON.stringify(formData),
            contentType: 'application/json; charset=utf-8',
            dataType    : 'json', // what type of data do we expect back from the server
            encode          : true,
            success: function (){
                $('#div-'.concat(divId.toString())).load(document.URL +  ' #div-'.concat(divId.toString()));
            }
        })
            // using the done promise callback
            .done(function(formData) {

                // log data to the console so we can see
                console.log(formData);

                // here we will handle errors and validation messages
            });
        // stop the form from submitting the normal way and refreshing the page

        event.preventDefault();
    });
});