/**
 * Copyright 2020-2022 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react';

import { render, fireEvent, screen } from 'design/utils/testing';

import FormLogin, { Props } from './FormLogin';

test('primary username and password with mfa off', () => {
  const onLogin = jest.fn();

  render(<FormLogin {...props} onLogin={onLogin} />);

  // Test only user/pwd form was rendered.
  expect(screen.queryByTestId('userpassword')).toBeVisible();
  expect(screen.queryByTestId('mfa-select')).not.toBeInTheDocument();
  expect(screen.queryByTestId('sso-list')).not.toBeInTheDocument();
  expect(screen.queryByTestId('passwordless')).not.toBeInTheDocument();

  // Test correct fn was called.
  fireEvent.change(screen.getByPlaceholderText(/username/i), {
    target: { value: 'username' },
  });
  fireEvent.change(screen.getByPlaceholderText(/password/i), {
    target: { value: '123' },
  });

  fireEvent.click(screen.getByText(/sign in/i));

  expect(onLogin).toHaveBeenCalledWith('username', '123', '');
});

test('auth2faType: otp', () => {
  const onLogin = jest.fn();

  render(<FormLogin {...props} auth2faType="otp" onLogin={onLogin} />);

  // Rendering of mfa dropdown.
  expect(screen.getByTestId('mfa-select')).not.toBeEmptyDOMElement();

  // fill form
  fireEvent.change(screen.getByPlaceholderText(/username/i), {
    target: { value: 'username' },
  });
  fireEvent.change(screen.getByPlaceholderText(/password/i), {
    target: { value: '123' },
  });
  fireEvent.change(screen.getByPlaceholderText(/123 456/i), {
    target: { value: '456' },
  });
  fireEvent.click(screen.getByText(/sign in/i));

  expect(onLogin).toHaveBeenCalledWith('username', '123', '456');
});

test('auth2faType: webauthn', async () => {
  const onLoginWithWebauthn = jest.fn();

  render(
    <FormLogin
      {...props}
      auth2faType="webauthn"
      onLoginWithWebauthn={onLoginWithWebauthn}
    />
  );

  // Rendering of mfa dropdown.
  expect(screen.getByTestId('mfa-select')).not.toBeEmptyDOMElement();

  // fill form
  fireEvent.change(screen.getByPlaceholderText(/username/i), {
    target: { value: 'username' },
  });
  fireEvent.change(screen.getByPlaceholderText(/password/i), {
    target: { value: '123' },
  });

  fireEvent.click(screen.getByText(/sign in/i));
  expect(onLoginWithWebauthn).toHaveBeenCalledWith({
    username: 'username',
    password: '123',
  });
});

test('input validation error handling', async () => {
  const onLogin = jest.fn();
  const onLoginWithSso = jest.fn();
  const onLoginWithWebauthn = jest.fn();

  render(
    <FormLogin
      {...props}
      auth2faType="otp"
      onLogin={onLogin}
      onLoginWithSso={onLoginWithSso}
      onLoginWithWebauthn={onLoginWithWebauthn}
    />
  );

  fireEvent.click(screen.getByText(/sign in/i));

  expect(onLogin).not.toHaveBeenCalled();
  expect(onLoginWithSso).not.toHaveBeenCalled();
  expect(onLoginWithWebauthn).not.toHaveBeenCalled();

  expect(screen.getByText(/username is required/i)).toBeInTheDocument();
  expect(screen.getByText(/password is required/i)).toBeInTheDocument();
  expect(screen.getByText(/token is required/i)).toBeInTheDocument();
});

test('error rendering', () => {
  render(
    <FormLogin
      {...props}
      auth2faType="off"
      attempt={{
        isFailed: true,
        isProcessing: false,
        message: 'errMsg',
        isSuccess: false,
      }}
    />
  );

  expect(screen.getByText('errMsg')).toBeInTheDocument();
});

test('primary sso', () => {
  const onLoginWithSso = jest.fn();

  render(
    <FormLogin
      {...props}
      authProviders={[
        { name: 'github', type: 'github', url: '' },
        { name: 'google', type: 'saml', url: '' },
      ]}
      onLoginWithSso={onLoginWithSso}
      primaryAuthType="sso"
    />
  );

  // Test only sso form was rendered.
  expect(screen.queryByTestId('sso-list')).toBeVisible();
  expect(screen.queryByTestId('passwordless')).not.toBeInTheDocument();
  expect(screen.queryByTestId('userpassword')).not.toBeInTheDocument();

  // Test clicking calls the right fn.
  fireEvent.click(screen.getByText(/github/i));
  expect(onLoginWithSso).toHaveBeenCalledTimes(1);
});

test('primary passwordless', () => {
  const onLoginWithSso = jest.fn();

  render(
    <FormLogin
      {...props}
      onLoginWithSso={onLoginWithSso}
      primaryAuthType="passwordless"
    />
  );

  // Test only passwordless form was rendered.
  expect(screen.queryByTestId('passwordless')).toBeVisible();
  expect(screen.queryByTestId('sso-list')).not.toBeInTheDocument();
  expect(screen.queryByTestId('userpassword')).not.toBeInTheDocument();
});

const props: Props = {
  auth2faType: 'off',
  authProviders: [],
  attempt: {
    isFailed: false,
    isProcessing: false,
    message: '',
    isSuccess: false,
  },
  onLogin: null,
  onLoginWithSso: null,
  onLoginWithWebauthn: null,
  isPasswordlessEnabled: false,
  primaryAuthType: 'local',
  privateKeyPolicyEnabled: false,
};
