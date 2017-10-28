EXT_MOCKS := \
	dynamodbiface \
	s3iface
EXT_MOCKS_DIR := extmocks
MOCKS := \
	dynamodb/dynamodbiface \
	apigateway \
	apigateway/apigatewayiface \
	s3/s3iface

MOCK_TGTS :=
define MOCK_TGT_template
MOCK_TGTS += $(1)/mocks/mock.go
endef

$(foreach MOCK,$(MOCKS),$(eval $(call MOCK_TGT_template,$(MOCK))))

EXT_MOCK_TGTS :=
define EXT_MOCK_TGT_template
EXT_MOCK_TGTS += $(EXT_MOCKS_DIR)/$(1)/mocks/mock.go
endef

$(foreach MOCK,$(EXT_MOCKS),$(eval $(call EXT_MOCK_TGT_template,$(MOCK))))

clean-mocks: ## remove all mocks
	find . -type d -name "mocks" -exec rm -rf {} +
	rm -rf $(EXT_MOCKS_DIR)

mocks: mocks-internal mocks-external ## build all interface mocks

mocks-internal: $(MOCK_TGTS) ## build all interface mocks in the project

mocks-external: $(EXT_MOCK_TGTS) ## build all interface mocks from $GOPATH

dynamodb/dynamodbiface/mocks/mock.go: IFACES := Service
apigateway/apigatewayiface/mocks/mock.go: IFACES := Service
apigateway/mocks/mock.go: IFACES := HTTPClient
s3/s3iface/mocks/mock.go: IFACES := Service

%/mocks/mock.go:
	@mkdir -p $(@D)
	mockgen -destination $@ $(IMPORT_PATH)/$* $(IFACES)

$(EXT_MOCKS_DIR)/dynamodbiface/mocks/mock.go: IFACES := DynamoDBAPI
$(EXT_MOCKS_DIR)/dynamodbiface/mocks/mock.go: EXT_IMPORT_PATH := github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface

$(EXT_MOCKS_DIR)/s3iface/mocks/mock.go: IFACES := S3API
$(EXT_MOCKS_DIR)/s3iface/mocks/mock.go: EXT_IMPORT_PATH := github.com/aws/aws-sdk-go/service/s3/s3iface

$(EXT_MOCKS_DIR)/%/mocks/mock.go:
	@mkdir -p $(@D)
	mockgen -destination $@ $(EXT_IMPORT_PATH) $(IFACES)
