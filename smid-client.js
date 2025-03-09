const EXTENSION_ID = 'poohpalffkmnicfcmdjofoaeagghgmdc';

class SMIDClient {
  constructor() {
    this.isExtensionAvailable = null;
  }

  getExtensionVersion() {
    return new Promise((resolve, reject) => {
      chrome.runtime.sendMessage(
          EXTENSION_ID,
          { action: 'getVersion' },
          (response) => {
            if (chrome.runtime.lastError) {
              reject(new Error('Extension connection error: ' + chrome.runtime.lastError.message));
              return;
            }
            if (response && response.success) {
              resolve(response.version);
            } else {
              reject(new Error(response?.error || 'Request failed or was rejected'));
            }
          }
      );
    });
  }

  async checkExtension() {
    if (this.isExtensionAvailable !== null) {
      return this.isExtensionAvailable;
    }
    try {
      const version = await this.getExtensionVersion();
      this.isExtensionAvailable = !!version;
    } catch (error) {
      this.isExtensionAvailable = false;
    }
    return this.isExtensionAvailable;
  }

  getAuthCode(domain) {
    return new Promise(async (resolve, reject) => {
      try {
        const isAvailable = await this.checkExtension();
        if (!isAvailable) {
          reject(new Error('SMID extension not found. Please install the extension first.'));
          return;
        }
        if (!domain || typeof domain !== 'string' || !domain.endsWith('smartschool.be')) {
          reject(new Error('Invalid domain. Must be a valid *.smartschool.be domain.'));
          return;
        }
        chrome.runtime.sendMessage(
            EXTENSION_ID,
            { action: 'getAuthCode', domain: domain },
            (response) => {
              if (chrome.runtime.lastError) {
                reject(new Error('Extension connection error: ' + chrome.runtime.lastError.message));
                return;
              }
              if (response && response.success) {
                resolve(response.data);
              } else {
                reject(new Error(response?.error || 'Request failed or was rejected'));
              }
            }
        );
      } catch (error) {
        reject(new Error('SMID extension not found or connection error'));
      }
    });
  }
}

window.SMIDClient = SMIDClient;
