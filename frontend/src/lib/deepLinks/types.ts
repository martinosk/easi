export interface DeepLinkHandler {
  param: string;
  onFound: (id: string) => void;
  onNotFound?: () => void;
}

export interface DeepLinkParam {
  param: string;
  routes: readonly string[];
}
