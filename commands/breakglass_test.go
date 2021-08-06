package cmd

import (
	"testing"

	cst "thy/constants"
	"thy/errors"
	"thy/fake"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetBreakGlassGetStatusCmd(t *testing.T) {
	testCase := []struct {
		name        string
		args        string
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"success",
			"",
			[]byte(`test`),
			nil,
		},
	}

	_, err := GetBreakGlassGetStatusCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tc := range testCase {

		t.Run(tc.name, func(t *testing.T) {
			fakeOutC := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			fakeOutC.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.out, tc.expectedErr
			}

			c := breakGlass{req, fakeOutC}
			_ = c.handleBreakGlassGetStatusCmd([]string{tc.args})

			if tc.expectedErr == nil {
				assert.Equal(t, data, tc.out)
			} else {
				assert.Equal(t, err, tc.expectedErr)
			}
		})
	}
}

func TestGetBreakGlassGenerateCmd(t *testing.T) {
	type flag struct {
		flag, value string
	}

	testCase := []struct {
		name        string
		args        []string
		out         []byte
		expectedErr *errors.ApiError
		flags       []flag
	}{
		{
			"success",
			[]string{"--min-number-of-shares", "2", "--new-admins", "bguser1,bguser2,bguser3"},
			[]byte(`test`),
			nil,
			[]flag{
				{
					cst.NewAdmins,
					"bguser1,bguser2,bguser3",
				},
				{
					cst.MinNumberOfShares,
					"2",
				},
			},
		},
	}

	_, err := GetBreakGlassGenerateCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tc := range testCase {

		t.Run(tc.name, func(t *testing.T) {
			if tc.flags != nil {
				for _, f := range tc.flags {
					viper.Set(f.flag, f.value)
				}
			}

			fakeOutC := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			fakeOutC.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.out, tc.expectedErr
			}

			c := breakGlass{req, fakeOutC}
			_ = c.handleBreakGlassGenerateCmd(tc.args)

			if tc.expectedErr == nil {
				assert.Equal(t, data, tc.out)
			} else {
				assert.Equal(t, err, tc.expectedErr)
			}
		})
	}
}

func TestGetBreakGlassApplyCmd(t *testing.T) {
	type flag struct {
		flag, value string
	}

	testCase := []struct {
		name        string
		args        []string
		out         []byte
		expectedErr *errors.ApiError
		flags       []flag
	}{
		{
			"success",
			[]string{"--shares", "6lFNUss5WgccrKLH39oeO4gQ5c7kA1McXlhDZn6joXQ=Ncc9-J7XRm78c_4SVwQgBAS1_7O6u9rRPHvUETnTBfw=Kmsl6oh1IhdK5SC5J3q1FaMhZhsQvo-sCS3X1Rtln_g=NOdvmZtLRVSkyujYZWgDbq5SjMSrsRbK2ocJFLotMeE=,45pPuy9V9V5zKdF852RNJy9hDZtB02nL6BBzGETteb4=IlyZoX1GL8pBFlNEXeJP8SQfeAxGWg168Xxus6bMp8k=V0d43eNG4aqq8AlerGnDKfftL9x1DJ6eihMaWqeIt0U=r2GibR5fnloRcnS0Ly1zoqpCvv72OLlRkdIwsR09fek="},
			[]byte(`test`),
			nil,
			[]flag{
				{
					cst.Shares,
					"6lFNUss5WgccrKLH39oeO4gQ5c7kA1McXlhDZn6joXQ=Ncc9-J7XRm78c_4SVwQgBAS1_7O6u9rRPHvUETnTBfw=Kmsl6oh1IhdK5SC5J3q1FaMhZhsQvo-sCS3X1Rtln_g=NOdvmZtLRVSkyujYZWgDbq5SjMSrsRbK2ocJFLotMeE=,45pPuy9V9V5zKdF852RNJy9hDZtB02nL6BBzGETteb4=IlyZoX1GL8pBFlNEXeJP8SQfeAxGWg168Xxus6bMp8k=V0d43eNG4aqq8AlerGnDKfftL9x1DJ6eihMaWqeIt0U=r2GibR5fnloRcnS0Ly1zoqpCvv72OLlRkdIwsR09fek=",
				},
			},
		},
	}

	_, err := GetBreakGlassApplyCmd()
	assert.Nil(t, err)

	viper.Set(cst.Version, "v1")
	for _, tc := range testCase {

		t.Run(tc.name, func(t *testing.T) {
			if tc.flags != nil {
				for _, f := range tc.flags {
					viper.Set(f.flag, f.value)
				}
			}

			fakeOutC := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			fakeOutC.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			req := &fake.FakeClient{}
			req.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.out, tc.expectedErr
			}

			c := breakGlass{req, fakeOutC}
			_ = c.handleBreakGlassApplyCmd(tc.args)

			if tc.expectedErr == nil {
				assert.Equal(t, data, tc.out)
			} else {
				assert.Equal(t, err, tc.expectedErr)
			}
		})
	}
}
