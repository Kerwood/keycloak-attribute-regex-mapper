package kerwood;

import org.keycloak.models.ClientSessionContext;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.ProtocolMapperModel;
import org.keycloak.models.UserModel;
import org.keycloak.models.UserSessionModel;
import org.keycloak.protocol.oidc.mappers.AbstractOIDCProtocolMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAccessTokenMapper;
import org.keycloak.protocol.oidc.mappers.OIDCAttributeMapperHelper;
import org.keycloak.protocol.oidc.mappers.OIDCIDTokenMapper;
import org.keycloak.protocol.oidc.mappers.UserInfoTokenMapper;
import org.keycloak.provider.ProviderConfigProperty;
import org.keycloak.representations.IDToken;

import java.util.ArrayList;
import java.util.List;

public class UserAttributeMapper extends AbstractOIDCProtocolMapper
    implements OIDCAccessTokenMapper, OIDCIDTokenMapper, UserInfoTokenMapper {

  private static final List<ProviderConfigProperty> configProperties = new ArrayList<>();

  public static final String PROVIDER_ID = "user-attribute-filter";

  static final String USER_ATTR_NAME = "userAttributeName";
  static final String REGEX_FILTER = "regexFilter";

  static {
    configProperties.add(
        new ProviderConfigProperty(USER_ATTR_NAME, "User Attribute Name", "The user attribute to filter.",
            ProviderConfigProperty.STRING_TYPE, null));

    configProperties.add(
        new ProviderConfigProperty(REGEX_FILTER, "Regex Filter", "Regular Expression Filter for the user attributes.",
            ProviderConfigProperty.STRING_TYPE, null));

    OIDCAttributeMapperHelper.addTokenClaimNameConfig(configProperties);
    OIDCAttributeMapperHelper.addIncludeInTokensConfig(configProperties, UserAttributeMapper.class);
  }

  @Override
  public String getDisplayCategory() {
    return "Token mapper";
  }

  @Override
  public String getDisplayType() {
    return "User Attribute Filter";
  }

  @Override
  public String getHelpText() {
    return "Map a custom user attribute to a token claim with regex-based filtering.";
  }

  @Override
  public List<ProviderConfigProperty> getConfigProperties() {
    return configProperties;
  }

  @Override
  public String getId() {
    return PROVIDER_ID;
  }

  @Override
  protected void setClaim(final IDToken token,
      final ProtocolMapperModel mappingModel,
      final UserSessionModel userSession,
      final KeycloakSession keycloakSession,
      final ClientSessionContext clientSessionCtx) {

    mappingModel.getConfig().put("multivalued", "true");

    UserModel user = userSession.getUser();

    String userAttribute = mappingModel.getConfig().get(USER_ATTR_NAME);
    List<String> attributes = user.getAttributes().get(userAttribute);
    if (attributes == null) {
      attributes = List.of();
    }

    String regex = mappingModel.getConfig().get(REGEX_FILTER);
    if (regex != null && !regex.isEmpty()) {
      attributes = attributes.stream().filter(val -> val.matches(regex)).toList();
    }

    OIDCAttributeMapperHelper.mapClaim(token, mappingModel, attributes);
  }

}
