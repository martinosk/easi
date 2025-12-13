export interface InitiateLoginRequest {
  email: string;
}

export interface InitiateLoginResponse {
  authorizationUrl: string;
  _links: {
    authorize: string;
  };
}
