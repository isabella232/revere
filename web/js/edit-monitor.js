$(document).ready(function() {
  var $triggers = $('.trigger');
  // Don't allow the removal of triggers if there is only one
  if ($triggers.length == 1) {
    $('.remove-trigger').hide();
  }

  /*
   * Event handlers
   */
  $('#add-trigger').click(function(e) {
    e.preventDefault();
    $('.remove-trigger').show();
    $triggers.first()
      .clone()
      .insertAfter('.trigger:last')
      .find('input[type="text"]')
      .val('');
  });

  $(document.body).on('click', '.remove-trigger', function(e) {
    e.preventDefault();
    if ($('.trigger').length > 1) {
      $(this).parents('.trigger').remove();
    }
    if ($('.trigger').length == 1) {
      $('.remove-trigger').hide();
    }
  });

  $('#monitor-form').submit(function(e) {
    e.preventDefault();
    var $form = $(this);

    var url = $form.attr("action"),
      data = $.extend(
        getMonitorData(),
        getProbeData(),
        {"triggers": getTriggerData()});

    $.post(url, data, function(data) {
      console.log(data);
    }, "json")
      .fail(function() {
        console.log("Fail");
      });
  });
});

var getMonitorData = function() {
  return $('#monitor-info').find(':input').serializeObject();
}

var getProbeData = function() {
  return $('#probe').find(':input').serializeObject();
}

var getTriggerData = function() {
  var data = []
  $.each($('.trigger'), function() {
    var triggerInputs = $(this).find(':input').serializeObject();
    data.push(triggerInputs);
  });
  return data;
}
