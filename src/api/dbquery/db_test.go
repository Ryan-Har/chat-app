package dbquery

import (
	"testing"

	"github.com/Ryan-Har/chat-app/src/api/mocks"
	"github.com/golang/mock/gomock"
)

func TestAddExternalUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	exampleName := "John Doe"
	exampleIP := "192.168.1.1"
	mockDBQuery.EXPECT().AddExternalUser(exampleName, exampleIP).Return([][]any{}, nil)

	_, err := mockDBQuery.AddExternalUser(exampleName, exampleIP)

	if err != nil {
		t.Errorf("Unexpected error during AddExternalUser: %v", err)
	}
}

func TestGetExternalUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	exampleName := "John Doe"
	exampleIP := "192.168.1.1"
	mockDBQuery.EXPECT().GetExternalUser(exampleName, exampleIP).Return([][]any{}, nil)

	_, err := mockDBQuery.GetExternalUser(exampleName, exampleIP)

	if err != nil {
		t.Errorf("Unexpected error during GetExternalUser: %v", err)
	}
}

func TestGetExternalUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	var exampleid int64 = 52
	mockDBQuery.EXPECT().GetExternalUserByID(exampleid).Return([][]any{}, nil)

	_, err := mockDBQuery.GetExternalUserByID(exampleid)

	if err != nil {
		t.Errorf("Unexpected error during GetExternalUserByID: %v", err)
	}
}

func TestChatStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	exampleUuid := "d935fb72-796d-4418-8a36-bc228d143790"
	exampleTime := "2024-02-20 15:50:20.123456"
	mockDBQuery.EXPECT().ChatStart(exampleUuid, exampleTime).Return([][]any{}, nil)

	_, err := mockDBQuery.ChatStart(exampleUuid, exampleTime)

	if err != nil {
		t.Errorf("Unexpected error during GetExternalUserByID: %v", err)
	}
}

func TestChatEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	exampleUuid := "d935fb72-796d-4418-8a36-bc228d143790"
	exampleTime := "2024-02-20 15:50:20.123456"
	mockDBQuery.EXPECT().ChatEnd(exampleUuid, exampleTime).Return([][]any{}, nil)

	_, err := mockDBQuery.ChatEnd(exampleUuid, exampleTime)

	if err != nil {
		t.Errorf("Unexpected error during GetExternalUserByID: %v", err)
	}
}

func TestAddInternalUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	var exampleid int64 = 1
	exampleString := "John"
	mockDBQuery.EXPECT().AddInternalUser(exampleid, exampleString, exampleString, exampleString, exampleString).Return([][]any{}, nil)

	_, err := mockDBQuery.AddInternalUser(exampleid, exampleString, exampleString, exampleString, exampleString)

	if err != nil {
		t.Errorf("Unexpected error during AddInternalUser: %v", err)
	}
}

func TestGetInternalUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	var exampleid int64 = 52
	mockDBQuery.EXPECT().GetInternalUserByID(exampleid).Return([][]any{}, nil)

	_, err := mockDBQuery.GetInternalUserByID(exampleid)

	if err != nil {
		t.Errorf("Unexpected error during GetInternalUserByID: %v", err)
	}
}

func TestUpdateInternalUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	var exampleid int64 = 1
	exampleString := "John"
	mockDBQuery.EXPECT().UpdateInternalUserByID(exampleid, exampleid, exampleString, exampleString, exampleString, exampleString).Return([][]any{}, nil)

	_, err := mockDBQuery.UpdateInternalUserByID(exampleid, exampleid, exampleString, exampleString, exampleString, exampleString)

	if err != nil {
		t.Errorf("Unexpected error during UpdateInternalUserByID: %v", err)
	}
}

func TestAddMessageByUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	var exampleid int64 = 1
	exampleUuid := "d935fb72-796d-4418-8a36-bc228d143790"
	exampleTime := "2024-02-20 15:50:20.123456"
	exampleString := "John"
	mockDBQuery.EXPECT().AddMessageByUUID(exampleUuid, exampleid, exampleString, exampleTime).Return([][]any{}, nil)

	_, err := mockDBQuery.AddMessageByUUID(exampleUuid, exampleid, exampleString, exampleTime)

	if err != nil {
		t.Errorf("Unexpected error during AddMessageByUUID: %v", err)
	}

}

func TestGetAllMessagesByUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	exampleUuid := "d935fb72-796d-4418-8a36-bc228d143790"
	mockDBQuery.EXPECT().GetAllMessagesByUUID(exampleUuid).Return([][]any{}, nil)

	_, err := mockDBQuery.GetAllMessagesByUUID(exampleUuid)

	if err != nil {
		t.Errorf("Unexpected error during GetAllMessagesByUUID: %v", err)
	}
}

func TestGetAllChatsInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDBQuery := mock_dbquery.NewMockDBQueryHandler(ctrl)

	mockDBQuery.EXPECT().GetAllChatsInProgress().Return([][]any{}, nil)

	_, err := mockDBQuery.GetAllChatsInProgress()

	if err != nil {
		t.Errorf("Unexpected error during GetAllChatsInProgress: %v", err)
	}
}
