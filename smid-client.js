const EXTENSION_ID = 'poohpalffkmnicfcmdjofoaeagghgmdc';

class SMIDClient {
  constructor() {
    this.isExtensionAvailable = null;
  }

  async checkExtension() {
    try {
      if (this.isExtensionAvailable !== null) {
        return this.isExtensionAvailable;
      }

      return new Promise((resolve) => {
        try {
          chrome.runtime.sendMessage(
            EXTENSION_ID,
            { action: 'ping' },
            (response) => {
              const isAvailable = !chrome.runtime.lastError;
              this.isExtensionAvailable = isAvailable;
              resolve(isAvailable);
            }
          );

          setTimeout(() => {
            if (this.isExtensionAvailable === null) {
              this.isExtensionAvailable = false;
              resolve(false);
            }
          }, 500);
        } catch (e) {
          this.isExtensionAvailable = false;
          resolve(false);
        }
      });
    } catch (error) {
      this.isExtensionAvailable = false;
      return false;
    }
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
          {
            action: 'getAuthCode',
            domain: domain
          },
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
