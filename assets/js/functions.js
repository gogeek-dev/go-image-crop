jQuery(document).ready(function () {
	/* When click on change profile pic */

	jQuery('#change-profile-pic').on('click', function (e) {
		var name = $('#cropperImageUpload').val();

		// Check if empty of not
		if (name.length < 1) {
			alert('Please select your file ');
			return false;
		}
		else {
			jQuery('#profile_pic_modal').modal({ show: true });
		}
	});


});