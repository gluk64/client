package engine

import (
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/libkb/kex"
)

type Sibkey struct {
	KexCom
	secretPhrase string
	libkb.Contextified
}

// NewSibkey creates a sibkey add engine.
// The secretPhrase is needed before this engine can run because
// the weak id used in receive() is based on it.
func NewSibkey(g *libkb.GlobalContext, secretPhrase string) *Sibkey {
	return &Sibkey{
		secretPhrase: secretPhrase,
		Contextified: libkb.NewContextified(g),
	}
}

func (k *Sibkey) Name() string {
	return "Sibkey"
}

func (k *Sibkey) GetPrereqs() EnginePrereqs {
	return EnginePrereqs{Session: true}
}

func (k *Sibkey) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{libkb.SecretUIKind}
}

func (k *Sibkey) SubConsumers() []libkb.UIConsumer {
	return nil
}

// Run starts the engine.
func (k *Sibkey) Run(ctx *Context, args, reply interface{}) error {
	k.engctx = ctx
	k.server = kex.NewSender(kex.DirectionXtoY)

	var err error
	k.user, err = libkb.LoadMe(libkb.LoadUserArg{PublicKeyOptional: true})
	if err != nil {
		return err
	}

	dp := k.G().Env.GetDeviceID()
	if dp == nil {
		return libkb.ErrNoDevice
	}
	k.deviceID = *dp

	k.deviceSibkey, err = k.user.GetComputedKeyFamily().GetSibkeyForDevice(k.deviceID)
	if err != nil {
		k.G().Log.Warning("Sibkey.Run: error getting device sibkey: %s", err)
		return err
	}
	arg := libkb.SecretKeyArg{
		DeviceKey: true,
		Reason:    "new device install",
		Ui:        ctx.SecretUI,
		Me:        k.user,
	}
	k.sigKey, err = k.G().Keyrings.GetSecretKey(arg)
	if err != nil {
		k.G().Log.Warning("Sibkey.Run: GetSecretKey error: %s", err)
		return err
	}

	id, err := k.wordsToID(k.secretPhrase)
	if err != nil {
		return err
	}
	k.sessionID = id

	m := kex.NewMeta(k.user.GetUid(), id, libkb.DeviceID{}, k.deviceID, kex.DirectionYtoX)
	k.receive(m, kex.DirectionYtoX)
	return nil
}
