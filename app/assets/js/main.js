var loc = location.href.split('/')[3];

$(function() {
	switch (loc) 
	{
		case 'settings':
			EnableTwoFa();
			DisableTwoFa();
			VerifyTwoFa();
			CancelTwoFa();
			break;
		case 'users':
			ChangeUserAdminSetting();
			FakeCheckboxs();
			break;
		default:
			break;
	}
});
