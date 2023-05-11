package com.signalfx.agent.jmx;

import org.apache.commons.lang3.StringUtils;

import javax.security.auth.callback.*;
import javax.security.sasl.RealmCallback;
import java.util.logging.Logger;

public class ClientCallbackHandler implements CallbackHandler {
    private final String username;
    private char[] password;
    private final String realm;

    public ClientCallbackHandler(String username, String password, String realm) {
        this.username = username;
        if (password != null) {
            this.password = password.toCharArray();
        }
        this.realm = realm;
    }

    @Override
    public void handle(Callback[] callbacks) throws UnsupportedCallbackException {
        for (Callback callback : callbacks) {
            if (callback instanceof NameCallback) {
                ((NameCallback)callback).setName(this.username);
            } else if (callback instanceof PasswordCallback) {
                ((PasswordCallback)callback).setPassword(this.password);
            } else if (callback instanceof RealmCallback) {
                ((RealmCallback)callback).setText(this.realm);
            } else {
                throw new UnsupportedCallbackException(callback);
            }
        }
    }

    @Override
    protected void finalize() {
        if (this.password != null) {
            for (int i = 0; i < this.password.length ; i++)
                this.password[i] = 0;
            this.password = null;
        }
    }
}