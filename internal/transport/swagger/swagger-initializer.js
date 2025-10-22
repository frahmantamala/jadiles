window.onload = function () {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [
      {
        name: "v1",
        url: window.location.origin + "/openapi3.json",
      },
      {
        name: "v2",
        url: window.location.origin + "/v2_openapi3.json",
      },
      {
        name: "backyard v1",
        url: window.location.origin + "/backyard_openapi3.json",
      }
    ],
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
